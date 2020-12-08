从前面两节内容可对gRPC的服务发现与负载均衡的主要内容（gRPC 本身没有提供注册中心，但为开发者提供了实现注册中心的接口，开发者是要实现其接口），下面将对流程进行梳理。gRPC版本为1.14.0（与前面代码不同，但流程大体一致）

#### 1. 接口定义

**1.2.1 Resolver和Balancer**

 **Picker**接口，一般来说每个GRPC负载均衡器都会带一个Picker，其唯一的Pick方法用来根据一定的条件选取一个PickResult的连接(SubConn).在负载均衡器初始化时，以及连接状态变化时，会触发更新Picker.定义如下：

```go
// The pickers used by gRPC can be updated by ClientConn.UpdateState().
type Picker interface {
	// Pick returns the connection to use for this RPC and related information.
	//
	// Pick should not block.  If the balancer needs to do I/O or any blocking
	// or time-consuming work to service this call, it should return
	// ErrNoSubConnAvailable, and the Pick call will be repeated by gRPC when
	// the Picker is updated (using ClientConn.UpdateState).
	//
	// If an error is returned:
	//
	// - If the error is ErrNoSubConnAvailable, gRPC will block until a new
	//   Picker is provided by the balancer (using ClientConn.UpdateState).
	//
	// - If the error is a status error (implemented by the grpc/status
	//   package), gRPC will terminate the RPC with the code and message
	//   provided.
	//
	// - For all other errors, wait for ready RPCs will wait, but non-wait for
	//   ready RPCs will be terminated with this error's Error() string and
	//   status code Unavailable.
	Pick(info PickInfo) (PickResult, error)
}
```

**Balancer**是负载均衡器所需要实现的接口:

```go
// UpdateClientConnState, ResolverError, UpdateSubConnState, and Close are
// guaranteed to be called synchronously from the same goroutine.  There's no
// guarantee on picker.Pick, it may be called anytime.
type Balancer interface {
	// 当连接状态发生变化时会调用此方法, 要求负载均衡器更具状态变化更新自身的连接策略和连接池(如果有的话)
	// 同时也需要更新对应的Picker
	UpdateClientConnState(ClientConnState) error
	// ResolverError is called by gRPC when the name resolver reports an error.
	ResolverError(error)
	// UpdateSubConnState is called by gRPC when the state of a SubConn
	// changes.
	UpdateSubConnState(SubConn, SubConnState)
	// Close closes the balancer. The balancer is not required to call
	// ClientConn.RemoveSubConn for its existing SubConns.
	Close()
}
```

简单来说一个**Balancer**要做的就两件事情:

- ​    1.管理客户端连接(SubConn)，当连接状态变化时更新自身的连接策略和连接池(如果有的话)；
- ​    2.创建和维护Picker

**Resolver**可以用来获取和更新连接地址，特别的当连接地址需要通过ZK等的注册中心，或者一些第三方的负载均衡服务获取时，就可以通过定制Resolver来解析。

```go
type Builder interface {
	// 在Dial方法中会异步调用该方法，构造一个Resolver
	Build(target Target, cc ClientConn, opts BuildOptions) (Resolver, error)
	// Scheme returns the scheme supported by this resolver.
	// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
	Scheme() string
}

type Resolver interface {
	// 在GRPC建立新的Transport时，会触发调用这个方法，用来获取或者生成一批新的客户端地址
	// 何为Transport？后面的篇幅会有解释
	ResolveNow(ResolveNowOptions)
	// Close closes the resolver.
	Close()
}
```



#### 2. 源码+流程

 先从Dial方法入手，Dial方法调用了DialContext方法，该主要用来初始化一个ClientConn对象，略去源码中我们不关心的部分，将重点部分添加注释如下：

```go
func DialContext(ctx context.Context, target string, opts ...DialOption) (conn *ClientConn, err error) {
    // 初始化 ClientConn, 这里将ClientConn变量命名为cc, 记住这个名字, 后续会有很多地方出现
	cc := &ClientConn{
		target:            target,
		csMgr:             &connectivityStateManager{},
        // 初始化地址池
		conns:             make(map[*addrConn]struct{}),
		dopts:             defaultDialOptions(),
        // 初始化一个空的 PickerWrapper ，PickerWrapper内聚了Picker，但是Picker并不在此时创建
        // 这里使用了策略模式，将Picker的具体算法独立成接口，Golang GRPC中使用了大量类似的策略模式
		blockingpicker:    newPickerWrapper(),
		czData:            new(channelzData),
		firstResolveEvent: grpcsync.NewEvent(),
	}
	// ... 略去一坨proxy, agency, 超时控制等相关的代码

	if cc.dopts.resolverBuilder == nil {
		// ... 略去一坨 获取默认resolverBuilder的代码
		}
	} else {
		cc.parsedTarget = resolver.Target{Endpoint: target}
	}
	// ... 略去一坨auth相关的代码

	// 新建 ResolverWrapper 和 resolver. 这又是一个策略模式
	rWrapper, err := newCCResolverWrapper(cc, resolverBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to build resolver: %v", err)
	}
    // 启动ResolverWrapper内部的watcher协程 处理地址以及连接状态的变更
    // 现版本将start放到了build过程中
	cc.resolverWrapper.start()

	// ...阻塞模式下 阻塞等待连接完成的代码

	return cc, nil
}
```

由此可见，当使用非阻塞模式时，此方法并没有真正创建连接，只是把连接的上下文信息都设置好了。接下来再看看DialContext方法中调用的newCCResolverWrapper 方法：

```go
// newCCResolverWrapper uses the resolver.Builder to build a Resolver and
// returns a ccResolverWrapper object which wraps the newly built resolver.
func newCCResolverWrapper(cc *ClientConn, rb resolver.Builder) (*ccResolverWrapper, error) {
   ccr := &ccResolverWrapper{
      cc:   cc,
      done: grpcsync.NewEvent(),
   }

   var credsClone credentials.TransportCredentials
   if creds := cc.dopts.copts.TransportCredentials; creds != nil {
      credsClone = creds.Clone()
   }
   rbo := resolver.BuildOptions{
      DisableServiceConfig: cc.dopts.disableServiceConfig,
      DialCreds:            credsClone,
      CredsBundle:          cc.dopts.copts.CredsBundle,
      Dialer:               cc.dopts.copts.Dialer,
   }

   var err error
   // We need to hold the lock here while we assign to the ccr.resolver field
   // to guard against a data race caused by the following code path,
   // rb.Build-->ccr.ReportError-->ccr.poll-->ccr.resolveNow, would end up
   // accessing ccr.resolver which is being assigned here.
   ccr.resolverMu.Lock()
   defer ccr.resolverMu.Unlock()
   ccr.resolver, err = rb.Build(cc.parsedTarget, ccr, rbo)
   if err != nil {
      return nil, err
   }
   return ccr, nil
}
```

 可以看到这里通过调用resolverBuilder的Build方法，创建了Resolver，再来看resolverWrapper.start()方法，该方法启动了一个协程调用resolverWrapper的watcher：

```go
// watcher processes address updates and service config updates sequentially.
// Otherwise, we need to resolve possible races between address and service
// config (e.g. they specify different balancer types).
func (ccr *ccResolverWrapper) watcher() {
   for {
      select {
      case <-ccr.done:
         return
      default:
      }

      select {
      case addrs := <-ccr.addrCh:
         select {
         case <-ccr.done:
            return
         default:
         }
         grpclog.Infof("ccResolverWrapper: sending new addresses to cc: %v", addrs)
         ccr.cc.handleResolvedAddrs(addrs, nil)
      case sc := <-ccr.scCh:
         select {
         case <-ccr.done:
            return
         default:
         }
         grpclog.Infof("ccResolverWrapper: got new service config: %v", sc)
         ccr.cc.handleServiceConfig(sc)
      case <-ccr.done:
         return
      }
   }
}
```

 当有新的地址或者当有服务状态变更时，分别调用了ClientConn的handleResolvedAddrs和handleServiceConfig。首先来看看ccr.addrCh和ccr.scCh这两个通道的值什么时候插入，分析其调用链：

- ResolverBuilder.Build/Resolver的watcher Loop -> resolverWrapper.NewAddress -> add new addr to addrCh
- ResolverBuilder.Build/Resolver的watcher Loop -> resolverWrapper.NewServiceConfig -> add new addr to scCh

​    也即resolver获取到新的地址或者服务状态变更后，会将信息传递给resolverWrapper，resolverWrapper则进一步将信息传递给ClientConn处理，那么再来看ClientConn的handleResolvedAddrs和handleServiceConfig，先看handleResolvedAddrs：

```go
func (cc *ClientConn) handleResolvedAddrs(addrs []resolver.Address, err error) {
   cc.mu.Lock()
   defer cc.mu.Unlock()
   if cc.conns == nil {
     // cc was closed.
		return
	}
    
    // 如果当前地址和新地址一致 则什么都不做
	if reflect.DeepEqual(cc.curAddresses, addrs) {
		return
	}

	cc.curAddresses = addrs

	if cc.dopts.balancerBuilder == nil {
		// ... 省略一坨代码 当balancerBuilder为nil时 后去一个默认的
	} else if cc.balancerWrapper == nil {
		// 初始化balancerWrapper和balancer 又是类似的策略模式
      cc.balancerWrapper = newCCBalancerWrapper(cc, cc.dopts.balancerBuilder, cc.balancerBuildOpts)
   }

   cc.balancerWrapper.handleResolvedAddrs(addrs, nil)
}

func newCCBalancerWrapper(cc *ClientConn, b balancer.Builder, bopts balancer.BuildOptions) *ccBalancerWrapper {
	ccb := &ccBalancerWrapper{
		cc:               cc,
		stateChangeQueue: newSCStateUpdateBuffer(),
		resolverUpdateCh: make(chan *resolverUpdate, 1),
		done:             make(chan struct{}),
		subConns:         make(map[*acBalancerWrapper]struct{}),
	}
	go ccb.watcher()
	ccb.balancer = b.Build(ccb, bopts)
	return ccb
}
```

这里终于看到balancer的出场了，最终由resolver获取到的新地址交由balancerWrapper的handleResolvedAddrs处理：

```go
func (ccb *ccBalancerWrapper) handleResolvedAddrs(addrs []resolver.Address, err error) {
	select {
	case <-ccb.resolverUpdateCh:
	default:
	}
	ccb.resolverUpdateCh <- &resolverUpdate{
		addrs: addrs,
		err:   err,
	}
}
```

balancerWrapper只是简单的将获取到的地址信息丢入其内部的一个通道，对于这个通道的处理则在balancerWrapper的watcher方法中：

```go
func (ccb *ccBalancerWrapper) watcher() {
	for {
		select {
		case t := <-ccb.stateChangeQueue.get():
			ccb.stateChangeQueue.load()
			select {
			case <-ccb.done:
				ccb.balancer.Close()
				return
			default:
			}
			ccb.balancer.HandleSubConnStateChange(t.sc, t.state)
		case t := <-ccb.resolverUpdateCh:
			select {
			case <-ccb.done:
				ccb.balancer.Close()
				return
			default:
			}
			ccb.balancer.HandleResolvedAddrs(t.addrs, t.err)
		case <-ccb.done:
		}

		select {
		case <-ccb.done:
			ccb.balancer.Close()
			ccb.mu.Lock()
			scs := ccb.subConns
			ccb.subConns = nil
			ccb.mu.Unlock()
			for acbw := range scs {
				ccb.cc.removeAddrConn(acbw.getAddrConn(), errConnDrain)
			}
			return
		default:
		}
	}
}

```

这些代码是不是看起来很眼熟，没错和resolverWrapper的设计模式时一致的。对于接受到的新地址，最终会调用balancer.HandleResolvedAddrs，也即是交由各个负载均衡器处理。我们先来理一下这条调用路径：

- resolverWrapper.handleResolvedAddrs - balancerWrapper.handleResolvedAddrs - balancer.HandleResolvedAddrs

​    handleServiceConfig的调用路径也是类似的，这里就不再详细描述了：

- resolverWrapper.handleServiceConfig - balancerWrapper.handleResolvedAddrs - balancer.HandleSubConnStateChange

​    整理下通过resolver更新balancer地址池的调用路径：

<div align="center"> <img src="../../pics/764de918-7881-46cc-a0b7-b2373794c048.png" width="500px"> </div><br>

**1.2.2 Picker**

