<!-- GFM-TOC -->

* [一 、服务注册与发布概述](#一-服务注册与发布概述)
* [二、服务发现的路由方式](#二-服务发现的路由方式)

  - [1.客户端路由](#1-客户端路由)
  - [2.代理层路由](#1-代理层路由)
* [三、gRPC服务发现](#三-gRPC服务发现)
* [四、最终实现](#四-最终实现)
* [参考](#参考)

<!-- GFM-TOC -->

# 一、服务注册与发布概述

**1.1 解决的问题**

服务注册与发布主要解决的服务依赖问题,通常意义上,如果A服务调用B服务时,最直接的做法是配置IP地址和端口.但随着服务依赖变多时,配置将会十分庞杂,且当服务发生迁移时,那么所有相关服务的配置均需要修改,这将十分难以维护以及容易出现问题.因此为了解决这种服务依赖关系,服务注册与发布应运而生.

**1.2 机制**

<div align="center"> <img src="../../pics/1607311668770.png" width="500px"> </div><br>

服务注册与发现主要分为以下几点.

- 服务信息发布

  这里主要是服务的服务名,IP信息,以及一些附件元数据.通过注册接口注册到服务注册发布中心.

- 存活检测

  当服务意外停止时,客户端需要感知到服务停止,并将服务的IP地址踢出可用的IP地址列表,这里可以使用定时心跳去实现.

- 客户端负载均衡

  通过服务注册与发布,可以实现一个服务部署多台实例,客户端实现在实例直接的负载均衡,从而实现服务的横向扩展.

  因此,服务注册与发布可以概况为,服务将信息上报,客户端拉取服务信息,通过服务名进行调用,当服务宕机时客户端踢掉故障服务,服务新上线时客户端自动添加到调用列表.

  grpc-go的整个实现大量使用go的接口特性,因此通过扩展接口,可以很容易的实现服务的注册与发现,这里服务注册中心考虑到可用性以及一致性,一般采用etcd或zookeeper来实现,这里实现etcd的版本.

# 二、服务发现的路由方式

## 1. 客户端路由

客户端路由模式，也就是调用方负责获取被调用方的地址信息，并使用相应的负载均衡算法发起请求。调用方访问服务注册服务，获取对应的服务 IP 地址和端口，可能还包括对应的服务负载信息（负载均衡算法、服务实例权重等）。调用方通过负载均衡算法选取其中一个发起请求。如下：

<div align="center"> <img src="../../pics/16073138601806.png" width="500px"> </div><br>

## 2. 代理层路由

代理层路由，不是由调用方去获取被调方的地址，而是通过代理的方式，由代理去获取被调方的地址、发起调用请求。client 只是会对代理层发起简单请求，代理层去进行 server 寻址、负载均衡等。如下：

<div align="center"> <img src="../../pics/16073139192146.png" width="500px"> </div><br>

grpc 官方介绍的服务发现流程图可以看出，grpc 是使用客户端路由的方式：

1、启动时，grpc client 通过服名字解析服务得到一个 address list，每个 address 将指示它是服务器地址还是负载平衡器地址，以及指示要哪个客户端负载平衡策略的服务配置（例如，round_robin 或 grpclb）

2、客户端实例化负载均衡策略 如果解析程序返回的任何一个地址是负载均衡器地址，则无论 service config 中定义了什么负载均衡策略，客户端都将使用grpclb策略。否则，客户端将使用 service config 中定义的负载均衡策略。如果服务配置未请求负载均衡策略，则客户端将默认使用选择第一个可用服务器地址的策略。

3、负载平衡策略为每个服务器地址创建一个 subchannel，假如是 grpclb 策略，客户端会根据名字解析服务返回的地址列表，请求负载均衡器，由负载均衡器决定请求哪个 subConn，然后打开一个数据流，对这个 subConn 中的所有服务器 adress 都建立连接，从而实现 client stream 的效果

4、当有rpc请求时，负载均衡策略决定哪个子通道即grpc服务器将接收请求，当可用服务器为空时客户端的请求将被阻塞。

# 三、gRPC服务发现

在`grpc client`的`DialContext`的方法中，有这一段关于`resolver`的代码：

```go
// Determine the resolver to use.
	cc.parsedTarget = grpcutil.ParseTarget(cc.target, cc.dopts.copts.Dialer != nil)
	channelz.Infof(logger, cc.channelzID, "parsed scheme: %q", cc.parsedTarget.Scheme)
	resolverBuilder := cc.getResolver(cc.parsedTarget.Scheme)
	if resolverBuilder == nil {
		// If resolver builder is still nil, the parsed target's scheme is
		// not registered. Fallback to default resolver and set Endpoint to
		// the original target.
		channelz.Infof(logger, cc.channelzID, "scheme %q not registered, fallback to default scheme", cc.parsedTarget.Scheme)
		cc.parsedTarget = resolver.Target{
			Scheme:   resolver.GetDefaultScheme(),
			Endpoint: target,
		}
		resolverBuilder = cc.getResolver(cc.parsedTarget.Scheme)
		if resolverBuilder == nil {
			return nil, fmt.Errorf("could not get resolver for default scheme: %q", cc.parsedTarget.Scheme)
		}
	}
```

 这段代码主要干了两件事情，ParseTarget和 getResolver 获取了一个 resolverBuilder

ParseTarget其实就是将 target 赋值给了 resolver target 对象的 endpoint 属性，如下：

```go
func ParseTarget(target string, skipUnixColonParsing bool) (ret resolver.Target) {
	var ok bool
	ret.Scheme, ret.Endpoint, ok = split2(target, "://")
	if !ok {
		if strings.HasPrefix(target, "unix:") && !skipUnixColonParsing {
			// Handle the "unix:[path]" case, because splitting on :// only
			// handles the "unix://[/absolute/path]" case. Only handle if the
			// dialer is nil, to avoid a behavior change with custom dialers.
			return resolver.Target{Scheme: "unix", Endpoint: target[len("unix:"):]}
		}
		return resolver.Target{Endpoint: target}
	}
	ret.Authority, ret.Endpoint, ok = split2(ret.Endpoint, "/")
	if !ok {
		return resolver.Target{Endpoint: target}
	}
	if ret.Scheme == "unix" {
		// Add the "/" back in the unix case, so the unix resolver receives the
		// actual endpoint.
		ret.Endpoint = "/" + ret.Endpoint
	}
	return ret
}
```

 getResolver的resolver.Get 方法 ，这里从一个 map 中取出了一个 Builder

```go
var (
	// m is a map from scheme to resolver builder.
	m = make(map[string]Builder)
	// defaultScheme is the default scheme to use.
	defaultScheme = "passthrough"
)

// Get returns the resolver builder registered with the given scheme.
//
// If no builder is register with the scheme, nil will be returned.
func Get(scheme string) Builder {
	if b, ok := m[scheme]; ok {
		return b
	}
	return nil
}
```

**resolver**

```go
// Package resolver defines APIs for name resolution in gRPC.
// All APIs in this package are experimental.
```

resolver 主要提供了一个名字解析的规范，所有的名字解析服务可以实现这个规范，包括 dns 解析类 dns_resolver 就是实现了这个规范的一个解析器。

resolver 中定义了 Builder ，通过调用 Build 去获取一个 resolver 实例

```go
// Builder creates a resolver that will be used to watch name resolution updates.
type Builder interface {
    // Build creates a new resolver for the given target.
    //
    // gRPC dial calls Build synchronously, and fails if the returned error is
    // not nil.
    Build(target Target, cc ClientConn, opts BuildOption) (Resolver, error)
    // Scheme returns the scheme supported by this resolver.
    // Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
    Scheme() string
}
```

在调用 Dial 方法发起 rpc 请求之前需要创建一个 ClientConn 连接，在 DialContext 这个方法中对 ClientConn 各属性进行了赋值，其中有一行代码就完成了 build resolver 的工作。

```go
// Build the resolver.
rWrapper, err := newCCResolverWrapper(cc, resolverBuilder)

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

不出意料，我们之前通过 get 去获取了一个 Builder， 这里调用了 Builder 的 Build 方法产生一个 resolver。

**register**

上面我们说到了，resolver 通过 get 方法，根据一个 string key 去一个 builder map 中获取一个 builder，这个 map 在 resolver 中初始化如下，那么是怎么进行赋值的呢？

```go
var (
    // m is a map from scheme to resolver builder.
    m = make(map[string]Builder)
    // defaultScheme is the default scheme to use.
    defaultScheme = "passthrough"
)
```

猜测肯定会有一个服务注册的过程，果然看到了一个 Register 方法

```go
func Register(b Builder) {
    m[b.Scheme()] = b
}
```

所有的 resolver 实现类通过 Register 方法去实现 Builder 的注册，比如 grpc 提供的 dnsResolver 这个类中调用了 init 方法，在服务初始化时实现了 Builder 的注册

```go
func init() {
    resolver.Register(NewBuilder())
}
```

**获取服务地址**

resolver 和 builder 都是 interface，也就是说它们只是定义了一套规则。具体实现由实现他们的子类去完成。例如在 helloworld 例子中，默认是通过默认的 passthrough 这个 scheme 去获取的 passthroughResolver 和 passthroughBuilder，我们来看 passthroughBuilder 的 Build 方法返回了一个带有 address 的 resolver，这个地址就是 server 的地址列表。在 helloworld demo 中，就是 “localhost:50051”。

```go
func (*passthroughBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
    r := &passthroughResolver{
        target: target,
        cc:     cc,
    }
    r.start()
    return r, nil
}
func (r *passthroughResolver) start() {
    r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: r.target.Endpoint}}})
}
```

**dns_resolver**

grpc 支持自定义 resolver 实现服务发现。同时 grpc 官方提供了一个基于 dns 的服务发现 resolver，这就是 dns_resolver，dns_resolver 通过 Build() 创建一个 resolver 实例，具体看一下 Build() 方法：

```go
// Build creates and starts a DNS resolver that watches the name resolution of the target.
func (b *dnsBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
    host, port, err := parseTarget(target.Endpoint, defaultPort)
    if err != nil {
        return nil, err
    }
    // IP address.
    if net.ParseIP(host) != nil {
        host, _ = formatIP(host)
        addr := []resolver.Address{{Addr: host + ":" + port}}
        i := &ipResolver{
            cc: cc,
            ip: addr,
            rn: make(chan struct{}, 1),
            q:  make(chan struct{}),
        }
        cc.NewAddress(addr)
        go i.watcher()
        return i, nil
    }
    // DNS address (non-IP).
    ctx, cancel := context.WithCancel(context.Background())
    d := &dnsResolver{
        freq:                 b.minFreq,
        backoff:              backoff.Exponential{MaxDelay: b.minFreq},
        host:                 host,
        port:                 port,
        ctx:                  ctx,
        cancel:               cancel,
        cc:                   cc,
        t:                    time.NewTimer(0),
        rn:                   make(chan struct{}, 1),
        disableServiceConfig: opts.DisableServiceConfig,
    }
    if target.Authority == "" {
        d.resolver = defaultResolver
    } else {
        d.resolver, err = customAuthorityResolver(target.Authority)
        if err != nil {
            return nil, err
        }
    }
    d.wg.Add(1)
    go d.watcher()
    return d, nil
}
```

在 Build 方法中，我们没有看到对 server address 寻址的过程，仔细找找，发现了一个 watcher 方法，如下：

```go
go d.watcher()
```

看一下 watcher 方法，发现它其实是一个监控进程，顾名思义作用是监控我们产生的 resolver 的状态，这里使用了一个 for 循环无限监听，通过 chan 进行消息通知。

```go
  func (d *dnsResolver) watcher() {
        defer d.wg.Done()
        for {
            select {
            case <-d.ctx.Done():
                return
            case <-d.t.C:
            case <-d.rn:
                if !d.t.Stop() {
                    // Before resetting a timer, it should be stopped to prevent racing with
                    // reads on it's channel.
                    <-d.t.C
                }
            }
            result, sc := d.lookup()
            // Next lookup should happen within an interval defined by d.freq. It may be
            // more often due to exponential retry on empty address list.
            if len(result) == 0 {
                d.retryCount++
                d.t.Reset(d.backoff.Backoff(d.retryCount))
            } else {
                d.retryCount = 0
                d.t.Reset(d.freq)
            }
            d.cc.NewServiceConfig(sc)
            d.cc.NewAddress(result)
            // Sleep to prevent excessive re-resolutions. Incoming resolution requests
            // will be queued in d.rn.
            t := time.NewTimer(minDNSResRate)
            select {
            case <-t.C:
            case <-d.ctx.Done():
                t.Stop()
                return
            }
        }
    }
```

定位到里面的 lookup 方法，进入 lookup 方法，发现它调用了 lookupSRV 这个方法：

```go
func (d *dnsResolver) lookup() ([]resolver.Address, string) {
    newAddrs := d.lookupSRV()
    // Support fallback to non-balancer address.
    newAddrs = append(newAddrs, d.lookupHost()...)
    if d.disableServiceConfig {
        return newAddrs, ""
    }
    sc := d.lookupTXT()
    return newAddrs, canaryingSC(sc)
}
```

继续追踪，lookupSRV 这个方法最终其实调用了 go 源码包 net 包下的 的 lookupSRV 方法，这个方法实现了 dns 协议对指定的service服务，protocol协议以及name域名进行srv查询，返回server 的 address 列表。经过层层解剖，我们终于找到了返回 server 的 address list 的代码。

```go
    _, srvs, err := d.resolver.LookupSRV(d.ctx, "grpclb", "tcp", d.host)
    ...
    func (r *Resolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*SRV, err error) {
        return r.lookupSRV(ctx, service, proto, name)
    }
```

### 总结

总结一下， grpc 的服务发现，主要通过 resolver 接口去定义，支持业务自己实现服务发现的 resolver。 grpc 提供了默认的 passthrough_resolver，不进行地址解析，直接将 client 发起请求时指定的 address （例如 helloworld client 指定地址为 “localhost:50051” ）当成 server address。同时，假如业务使用 dns 进行服务发现，grpc 提供了 dns_resolver，通过对指定的service服务，protocol协议以及name域名进行srv查询，来返回 server 的 address 列表。