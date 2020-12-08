 grpc 负载均衡的实现，这里主要分为初始化 balancer 和寻址两步

#### 1. 初始化 balancer

介绍 grpc 服务发现时，我们知道了通过 dns_resolver 的 lookup 方法可以得到一个 []resolver.Address，resolver.Address 结构如下：

```go
type Address struct {
   // Addr is the server address on which a connection will be established.
   Addr string

   // ServerName is the name of this address.
   // If non-empty, the ServerName is used as the transport certification authority for
   // the address, instead of the hostname from the Dial target string. In most cases,
   // this should not be set.
   //
   // If Type is GRPCLB, ServerName should be the name of the remote load
   // balancer, not the name of the backend.
   //
   // WARNING: ServerName must only be populated with trusted values. It
   // is insecure to populate it with data from untrusted inputs since untrusted
   // values could be used to bypass the authority checks performed by TLS.
   ServerName string

   // Attributes contains arbitrary data about this address intended for
   // consumption by the load balancing policy.
   Attributes *attributes.Attributes

   // Type is the type of this address.
   //
   // Deprecated: use Attributes instead.
   Type AddressType

   // Metadata is the information associated with Addr, which may be used
   // to make load balancing decision.
   //
   // Deprecated: use Attributes instead.
   Metadata interface{}
}
```

可以看到每一个 resolver.Address 包括服务器地址、地址类型、负载均衡策略以及包含其他额外数据的一个结构 Metadata。所以 []resolver.Address 里面包含了服务器地址列表，那么拿到地址列表之后具体干了什么事呢？下面我们来看看

```go
	state, err := d.lookup()
		if err != nil {
			d.cc.ReportError(err)
		} else {
			d.cc.UpdateState(*state)
		}

func (ccr *ccResolverWrapper) UpdateState(s resolver.State) {
	if ccr.done.HasFired() {
		return
	}
	channelz.Infof(logger, ccr.cc.channelzID, "ccResolverWrapper: sending update to cc: %v", s)
	if channelz.IsOn() {
		ccr.addChannelzTraceEvent(s)
	}
	ccr.curState = s
	ccr.poll(ccr.cc.updateResolverState(ccr.curState, nil))
}
```

这里调用了`UpdateState`方法，在 `UpdateState`这个方法里面又调用了 `updateResolverState`这个方法，对负载均衡器的初始化就是在这个方法中进行的，追踪到`applyServiceConfigAndBalancer`方法如下：

```go
if cc.dopts.balancerBuilder == nil {
		// Only look at balancer types and switch balancer if balancer dial
		// option is not set.
		var newBalancerName string
		if cc.sc != nil && cc.sc.lbConfig != nil {
			newBalancerName = cc.sc.lbConfig.name
		} else {
			var isGRPCLB bool
			for _, a := range addrs {
				if a.Type == resolver.GRPCLB {
					isGRPCLB = true
					break
				}
			}
			if isGRPCLB {
				newBalancerName = grpclbName
			} else if cc.sc != nil && cc.sc.LB != nil {
				newBalancerName = *cc.sc.LB
			} else {
				newBalancerName = PickFirstBalancerName
			}
		}
		cc.switchBalancer(newBalancerName)
	} else if cc.balancerWrapper == nil {
		// Balancer dial option was set, and this is the first time handling
		// resolved addresses. Build a balancer with dopts.balancerBuilder.
		cc.curBalancerName = cc.dopts.balancerBuilder.Name()
		cc.balancerWrapper = newCCBalancerWrapper(cc, cc.dopts.balancerBuilder, cc.balancerBuildOpts)
	}
```

之前说到了我们在 dns_resolver 中查找 address 时是通过 grpclb 去进行查找的，所以它返回的 resolver 的策略就是 grpclb 策略。这里会进入到 switchBalancer 方法，我们来看看这个方法：

```go
// switchBalancer starts the switching from current balancer to the balancer
// with the given name.
//
// It will NOT send the current address list to the new balancer. If needed,
// caller of this function should send address list to the new balancer after
// this function returns.
//
// Caller must hold cc.mu.
func (cc *ClientConn) switchBalancer(name string) {
	...
	builder := balancer.Get(name)
	...
	cc.curBalancerName = builder.Name()
	cc.balancerWrapper = newCCBalancerWrapper(cc, builder, cc.balancerBuildOpts)
}
```

这里通过 grpclb 这个 name 去获取到了 grpclb 策略的一个 balancer 实现，然后调用了 newCCBalancerWrapper 这个方法，继续跟踪

```go
func newCCBalancerWrapper(cc *ClientConn, b balancer.Builder, bopts balancer.BuildOptions) *ccBalancerWrapper {
	ccb := &ccBalancerWrapper{
		cc:       cc,
		scBuffer: buffer.NewUnbounded(),
		done:     grpcsync.NewEvent(),
		subConns: make(map[*acBalancerWrapper]struct{}),
	}
	go ccb.watcher()
	ccb.balancer = b.Build(ccb, bopts)
	return ccb
}
```

来看一下这里的 Build 方法，去 grpclb 这个策略的实现类里面看，发现它返回了一个 lbBalancer 实例

```go
func (bb *baseBuilder) Build(cc balancer.ClientConn, opt balancer.BuildOptions) balancer.Balancer {
   bal := &baseBalancer{
      cc:            cc,
      pickerBuilder: bb.pickerBuilder,

      subConns: make(map[resolver.Address]balancer.SubConn),
      scStates: make(map[balancer.SubConn]connectivity.State),
      csEvltr:  &balancer.ConnectivityStateEvaluator{},
      config:   bb.config,
   }
   // Initialize picker to a picker that always returns
   // ErrNoSubConnAvailable, because when state of a SubConn changes, we
   // may call UpdateState with this picker.
   bal.picker = NewErrPicker(balancer.ErrNoSubConnAvailable)
   return bal
}
```

#### 2.寻址

helloworld demo 中 client 发送请求主要分为三步，对 balancer 的初始化其实是在第一步 grpc.Dial 时初始化 dialContext 时完成的。那么寻址过程，就是在第三步调用 sayHello 时完成的。

进入 sayHello ——> c.cc.Invoke ——> invoke ——> newClientStream 方法中，有下面一段代码：

```go
// Only this initial attempt has stats/tracing.
	// TODO(dfawley): move to newAttempt when per-attempt stats are implemented.
	if err := cs.newAttemptLocked(sh, trInfo); err != nil {
		cs.finish(err)
		return nil, err
	}
```

进入 newAttemptLocked 方法，如下：

```go
// newAttemptLocked creates a new attempt with a transport.
// If it succeeds, then it replaces clientStream's attempt with this new attempt.
func (cs *clientStream) newAttemptLocked(sh stats.Handler, trInfo *traceInfo) (retErr error) {
	newAttempt := &csAttempt{
		cs:           cs,
		dc:           cs.cc.dopts.dc,
		statsHandler: sh,
		trInfo:       trInfo,
	}
	defer func() {
		if retErr != nil {
			// This attempt is not set in the clientStream, so it's finish won't
			// be called. Call it here for stats and trace in case they are not
			// nil.
			newAttempt.finish(retErr)
		}
	}()

	if err := cs.ctx.Err(); err != nil {
		return toRPCErr(err)
	}

	ctx := cs.ctx
	if cs.cc.parsedTarget.Scheme == "xds" {
		// Add extra metadata (metadata that will be added by transport) to context
		// so the balancer can see them.
		ctx = grpcutil.WithExtraMetadata(cs.ctx, metadata.Pairs(
			"content-type", grpcutil.ContentType(cs.callHdr.ContentSubtype),
		))
	}
	t, done, err := cs.cc.getTransport(ctx, cs.callInfo.failFast, cs.callHdr.Method)
	...
}
```

发现它调用了 getTransport 方法，进入这个方法，我们找到了 pick 方法的调用

```go
func (cc *ClientConn) getTransport(ctx context.Context, failfast bool, method string) (transport.ClientTransport, func(balancer.DoneInfo), error) {
   t, done, err := cc.blockingpicker.pick(ctx, failfast, balancer.PickInfo{
      Ctx:            ctx,
      FullMethodName: method,
   })
   if err != nil {
      return nil, nil, toRPCErr(err)
   }
   return t, done, nil
}
```

pick 方法即是具体寻址的方法，仔细看 pick 方法，它先 Pick 获取了一个 SubConn，SubConn 结构体中包含了一个 address list，然后它会对每一个 address 都会发送 rpc 请求。

```go
// pick returns the transport that will be used for the RPC.
// It may block in the following cases:
// - there's no picker
// - the current picker returns ErrNoSubConnAvailable
// - the current picker returns other errors and failfast is false.
// - the subConn returned by the current picker is not READY
// When one of these situations happens, pick blocks until the picker gets updated.
func (pw *pickerWrapper) pick(ctx context.Context, failfast bool, info balancer.PickInfo) (transport.ClientTransport, func(balancer.DoneInfo), error) {
   var ch chan struct{}

   var lastPickErr error
   for {
      ...
      p := pw.picker
      pw.mu.Unlock()

      pickResult, err := p.Pick(info)

      if err != nil {
         if err == balancer.ErrNoSubConnAvailable {
            continue
         }
         if _, ok := status.FromError(err); ok {
            // Status error: end the RPC unconditionally with this status.
            return nil, nil, err
         }
         // For all other errors, wait for ready RPCs should block and other
         // RPCs should fail with unavailable.
         if !failfast {
            lastPickErr = err
            continue
         }
         return nil, nil, status.Error(codes.Unavailable, err.Error())
      }

      acw, ok := pickResult.SubConn.(*acBalancerWrapper)
      if !ok {
         logger.Error("subconn returned from pick is not *acBalancerWrapper")
         continue
      }
      if t, ok := acw.getAddrConn().getReadyTransport(); ok {
         if channelz.IsOn() {
            return t, doneChannelzWrapper(acw, pickResult.Done), nil
         }
         return t, pickResult.Done, nil
      }
      if pickResult.Done != nil {
         // Calling done with nil error, no bytes sent and no bytes received.
         // DoneInfo with default value works.
         pickResult.Done(balancer.DoneInfo{})
      }
      logger.Infof("blockingPicker: the picked transport is not ready, loop back to repick")
      // If ok == false, ac.state is not READY.
      // A valid picker always returns READY subConn. This means the state of ac
      // just changed, and picker will be updated shortly.
      // continue back to the beginning of the for loop to repick.
   }
}
```

它调用了 pickWrapper 中的 Pick 方法，在第一步初始化 balancer 的时候我们说到，它返回的其实是 lbBalancer 实例，所以这里去看 lbBalancer 实例的 Pick 实现：

```go
func (p *lbPicker) Pick(balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Layer one roundrobin on serverList.
	s := p.serverList[p.serverListNext]
	p.serverListNext = (p.serverListNext + 1) % len(p.serverList)

	// If it's a drop, return an error and fail the RPC.
	if s.Drop {
		p.stats.drop(s.LoadBalanceToken)
		return balancer.PickResult{}, status.Errorf(codes.Unavailable, "request dropped by grpclb")
	}

	// If not a drop but there's no ready subConns.
	if len(p.subConns) <= 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	// Return the next ready subConn in the list, also collect rpc stats.
	sc := p.subConns[p.subConnsNext]
	p.subConnsNext = (p.subConnsNext + 1) % len(p.subConns)
	done := func(info balancer.DoneInfo) {
		if !info.BytesSent {
			p.stats.failedToSend()
		} else if info.BytesReceived {
			p.stats.knownReceived()
		}
	}
	return balancer.PickResult{SubConn: sc, Done: done}, nil
}
```

可以看到这其实是一个轮询实现。用一个指针表示这次取的位置，取过之后就更新这个指针为下一位。Pick 的返回是一个 SubConn 结构，SubConn 里面就包含了 server 的地址列表，此时寻址就完成了。

#### 3. 发起request

寻址完成之后，我们得到了包含 server 地址列表的 SubConn，接下来是如何发送请求的呢？在 pick 方法中接着往下看，发现了这段代码。

```go
    acw, ok := subConn.(*acBalancerWrapper)
    if !ok {
        grpclog.Error("subconn returned from pick is not *acBalancerWrapper")
        continue
    }
    if t, ok := acw.getAddrConn().getReadyTransport(); ok {
        if channelz.IsOn() {
            return t, doneChannelzWrapper(acw, done), nil
        }
        return t, done, nil
    }
```

这段代码先将 SubConn 转换成了一个 acBalancerWrapper ，然后获取其中的 addrConn 对象，接着调用 getReadyTransport 方法，如下：

```go
// getReadyTransport returns the transport if ac's state is READY.
// Otherwise it returns nil, false.
// If ac's state is IDLE, it will trigger ac to connect.
func (ac *addrConn) getReadyTransport() (transport.ClientTransport, bool) {
	ac.mu.Lock()
	if ac.state == connectivity.Ready && ac.transport != nil {
		t := ac.transport
		ac.mu.Unlock()
		return t, true
	}
	var idle bool
	if ac.state == connectivity.Idle {
		idle = true
	}
	ac.mu.Unlock()
	// Trigger idle ac to connect.
	if idle {
		ac.connect()
	}
	return nil, false
}
```

getReadyTransport 这个方法返回一个 Ready 状态的网络连接，假如连接状态是 IDLE 状态，会调用 connect 方法去进行客户端连接，connect 方法如下：

```go
func (ac *addrConn) connect() error {
    ac.mu.Lock()
    if ac.state == connectivity.Shutdown {
        ac.mu.Unlock()
        return errConnClosing
    }
    if ac.state != connectivity.Idle {
        ac.mu.Unlock()
        return nil
    }
    // Update connectivity state within the lock to prevent subsequent or
    // concurrent calls from resetting the transport more than once.
    ac.updateConnectivityState(connectivity.Connecting)
    ac.mu.Unlock()
    // Start a goroutine connecting to the server asynchronously.
    go ac.resetTransport()
    return nil
}
```

通过 go ac.resetTransport() 这一行可以看到 connect 方法新起协程异步去与 server 建立连接。resetTransport 方法中有一行调用了 tryAllAddrs 方法，如下：

```go
    newTr, addr, reconnect, err := ac.tryAllAddrs(addrs, connectDeadline)
```

猜测是在这个方法中去轮询 address 与 每个 address 的 server 建立连接。

```go
func (ac *addrConn) tryAllAddrs(addrs []resolver.Address, connectDeadline time.Time) (transport.ClientTransport, resolver.Address, *grpcsync.Event, error) {
    for _, addr := range addrs {
        ...
        newTr, reconnect, err := ac.createTransport(addr, copts, connectDeadline)
        if err == nil {
            return newTr, addr, reconnect, nil
        }
        ac.cc.blockingpicker.updateConnectionError(err)
    }
    // Couldn't connect to any address.
    return nil, resolver.Address{}, nil, fmt.Errorf("couldn't connect to any address")
}
```

一看这个方法，果然如此，遍历所有地址，然后调用了 createTransport 方法，为每个地址的服务器建立连接，看到这里，我们也明白了 Stream 的实现。传统的 client 实现是对某个 server 地址发起 connect，Stream 的实质无非是对一批 server 的 address 进行轮询并建立 connect。

### 总结

grpc 负载均衡的实现是通过客户端路由的方式，先通过服务发现获取一个 resolver.Address 列表，resolver.Address 中包含了服务器地址和负载均衡服务名字，通过这个名字去初始化响应的 balancer，dns_resolver 中默认是使用的 grpclb 这个负载均衡器，寻址方式是轮询。通过调用 picker 去 生成一个 SubConn，SubConn 中包括服务器的地址列表，采用异步的方式对地址列表进行轮询，然后为每一个服务端地址都进行 connect 。