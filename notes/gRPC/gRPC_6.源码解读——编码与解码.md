<!-- GFM-TOC -->

* [一 、概述](#一-概述)
* [二、定义](#二-定义)

- [三、参考gPRC-go的example](#三-参考gPRC-go的example)
- [四、最终实现](#四-最终实现)
- [参考](#参考)

<!-- GFM-TOC -->

# 一、概述

一般的协议都会包括协议头和协议体，对于业务而言，一般只关心需要发送的业务数据。所以，协议头的内容一般是框架自动帮忙填充。将业务数据包装成指定协议格式的数据包就是编码的过程，从指定协议格式中的数据包中取出业务数据的过程就是解码的过程。

每个 rpc 框架基本都有自己的编解码器，下面我们就来说说 grpc 的编解码过程。

# 二、grpc 解码

从helloworld demo 中 server 的 main 函数入手

```go
func main() {
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    pb.RegisterGreeterServer(s, &server{})
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

在 s.Serve(lis) ——> s.handleRawConn(rawConn) —— > s.serveStreams(st) ——> s.handleStream(st, stream, s.traceInfo(st, stream)) ——> s.processUnaryRPC(t, stream, srv, md, trInfo) 方法中有一段代码：

```go
sh := s.opts.statsHandler
...
df := func(v interface{}) error {
        if err := s.getCodec(stream.ContentSubtype()).Unmarshal(d, v); err != nil {
            return status.Errorf(codes.Internal, "grpc: error unmarshalling request: %v", err)
        }
        if sh != nil {
            sh.HandleRPC(stream.Context(), &stats.InPayload{
                RecvTime:   time.Now(),
                Payload:    v,
                WireLength: payInfo.wireLength,
                Data:       d,
                Length:     len(d),
            })
        }
        if binlog != nil {
            binlog.Log(&binarylog.ClientMessage{
                Message: d,
            })
        }
        if trInfo != nil {
            trInfo.tr.LazyLog(&payload{sent: false, msg: v}, true)
        }
        return nil
}
```

这段代码的逻辑先调 getCodec 获取解包类，然后调用这个类的 Unmarshal 方法进行解包。将业务数据取出来，然后调用 handler 进行处理。

```go
func (s *Server) getCodec(contentSubtype string) baseCodec {
    if s.opts.codec != nil {
        return s.opts.codec
    }
    if contentSubtype == "" {
        return encoding.GetCodec(proto.Name)
    }
    codec := encoding.GetCodec(contentSubtype)
    if codec == nil {
        return encoding.GetCodec(proto.Name)
    }
    return codec
}
```

来看 getCodec 这个方法，它是通过 contentSubtype 这个字段来获取解包类的。假如不设置 contentSubtype ，那么默认会用名字为 proto 的解码器。

我们来看看 contentSubtype 是如何设置的。之前说到了 grpc 的底层默认是基于 http2 的。在 serveHttp 时调用了 NewServerHandlerTransport 这个方法来创建一个 ServerTransport，然后我们发现，其实就是根据 content-type 这个字段去生成的。

```go
func NewServerHandlerTransport(w http.ResponseWriter, r *http.Request, stats stats.Handler) (ServerTransport, error) {
    ...
    contentType := r.Header.Get("Content-Type")
    // TODO: do we assume contentType is lowercase? we did before
    contentSubtype, validContentType := contentSubtype(contentType)
    if !validContentType {
        return nil, errors.New("invalid gRPC request content-type")
    }
    if _, ok := w.(http.Flusher); !ok {
        return nil, errors.New("gRPC requires a ResponseWriter supporting http.Flusher")
    }
    st := &serverHandlerTransport{
        rw:             w,
        req:            r,
        closedCh:       make(chan struct{}),
        writes:         make(chan func()),
        contentType:    contentType,
        contentSubtype: contentSubtype,
        stats:          stats,
    }
}
```

来看看 contentSubtype 这个方法 。

```go
...
baseContentType = "application/grpc"
...
func contentSubtype(contentType string) (string, bool) {
    if contentType == baseContentType {
        return "", true
    }
    if !strings.HasPrefix(contentType, baseContentType) {
        return "", false
    }
    // guaranteed since != baseContentType and has baseContentType prefix
    switch contentType[len(baseContentType)] {
    case '+', ';':
        // this will return true for "application/grpc+" or "application/grpc;"
        // which the previous validContentType function tested to be valid, so we
        // just say that no content-subtype is specified in this case
        return contentType[len(baseContentType)+1:], true
    default:
        return "", false
    }
}
```

可以看到 grpc 协议默认以 application/grpc 开头，假如不以这个开头会返回错误，假如我们想使用 json 的解码器，应该设置 content-type = application/grpc+json 。下面是一个基于 grpc 协议的请求 request ：

```http
HEADERS (flags = END_HEADERS)
:method = POST
:scheme = http
:path = /google.pubsub.v2.PublisherService/CreateTopic
:authority = pubsub.googleapis.com
grpc-timeout = 1S
content-type = application/grpc+proto
grpc-encoding = gzip
authorization = Bearer y235.wef315yfh138vh31hv93hv8h3v
DATA (flags = END_STREAM)
<Length-Prefixed Message>
```

详细可参考 [proto-http2](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md)

怎么拿的呢，再看一下 encoding.getCodec 方法

```go
func GetCodec(contentSubtype string) Codec {
    return registeredCodecs[contentSubtype]
}
```

它其实取得是 registeredCodecs 这个 map 中的 codec，这个 map 是 RegisterCodec 方法注册进去的。

```go
var registeredCodecs = make(map[string]Codec)
func RegisterCodec(codec Codec) {
    if codec == nil {
        panic("cannot register a nil Codec")
    }
    if codec.Name() == "" {
        panic("cannot register Codec with empty string result for Name()")
    }
    contentSubtype := strings.ToLower(codec.Name())
    registeredCodecs[contentSubtype] = codec
}
```

毫无疑问， encoding 目录的 proto 包下肯定在初始化时调用注册方法了。果然

```go
func init() {
    encoding.RegisterCodec(codec{})
}
```

绕了一圈，调用的其实是 proto 的 Unmarshal 方法，如下：

```go
func (codec) Unmarshal(data []byte, v interface{}) error {
    protoMsg := v.(proto.Message)
    protoMsg.Reset()
    if pu, ok := protoMsg.(proto.Unmarshaler); ok {
        // object can unmarshal itself, no need for buffer
        return pu.Unmarshal(data)
    }
    cb := protoBufferPool.Get().(*cachedProtoBuffer)
    cb.SetBuf(data)
    err := cb.Unmarshal(protoMsg)
    cb.SetBuf(nil)
    protoBufferPool.Put(cb)
    return err
}
```





在 gRPC 中，大类可分为两种 RPC 方法，与拦截器的对应关系是：

- 普通方法：一元拦截器（grpc.UnaryInterceptor）
- 流方法：流拦截器（grpc.StreamInterceptor）

 helloworld demo 客户端的 main 函数，grpc.Dial —> DialContext —> chainUnaryClientInterceptors

```go
// chainUnaryClientInterceptors chains all unary client interceptors into one.
// chainUnaryClientInterceptors 将所有的拦截器串接成一个拦截器
func chainUnaryClientInterceptors(cc *ClientConn) {
	interceptors := cc.dopts.chainUnaryInts
	// Prepend dopts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
	if cc.dopts.unaryInt != nil {
		interceptors = append([]UnaryClientInterceptor{cc.dopts.unaryInt}, interceptors...)
	}
	var chainedInt UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, invoker UnaryInvoker, opts ...CallOption) error {
			return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker), opts...)
		}
	}
	cc.dopts.unaryInt = chainedInt
}
```

- ctx context.Context：请求上下文
- req interface{}：RPC 方法的请求参数
- info *UnaryServerInfo：RPC 方法的所有信息
- handler UnaryHandler：RPC 方法本身

接下里查看`getChainUnaryInvoker`函数：

```go
// getChainUnaryInvoker recursively generate the chained unary invoker.
func getChainUnaryInvoker(interceptors []UnaryClientInterceptor, curr int, finalInvoker UnaryInvoker) UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) error {
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker), opts...)
	}
}
// 递归调用返回 UnaryInvoker 结构体，在UnaryInvoker实例化后会去调用第curr + 1个interceptors直至结束。
// interceptor0-interceptor1-interceptor2-interceptor3-...-interceptorn-finalinvoke
// 拦截器链会以递归遍历切片的方式递归调用所有拦截器，使请求依次通过这些拦截器，直到请求经过了所有的拦截器，
// 最终抵达服务端请求处理的接口函数，处理完之后返回。
// UnaryInvoker is called by UnaryClientInterceptor to complete RPCs.
type UnaryInvoker func(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) error
// 这里的 opts ...CallOption 参数：functional options API（详见基础）
```

返回值赋给`cc.dopts.unaryInt`,但没有立刻被调用。

```go
// 客户端调用SayHello
err := c.cc.Invoke(ctx, "/helloworld.Greeter/SayHello", in, out, opts...)
```

在这里的`Invoke`进行调用

```go
// Invoke sends the RPC request on the wire and returns after response is
// received.  This is typically called by generated code.
//
// All errors returned by Invoke are compatible with the status package.
func (cc *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...CallOption) error {
	// allow interceptor to see all applicable call options, which means those
	// configured as defaults from dial option as well as per-call options
	opts = combine(cc.dopts.callOptions, opts)

	if cc.dopts.unaryInt != nil {
    // 这里是调用的入口，递归调用，一直到最后
		return cc.dopts.unaryInt(ctx, method, args, reply, cc, invoke, opts...)
	}
	return invoke(ctx, method, args, reply, cc, opts...)
}
```

# 三、grpc 编码

在剖析解码代码的基础上，编码代码就很轻松了，其实直接找到 encoding 目录的 proto 包，看 Marshal 方法在哪儿被调用就行了。

于是我们很快就找到了调用路径，也是这个路径：

s.Serve(lis) ——> s.handleRawConn(rawConn) —— > s.serveStreams(st) ——> s.handleStream(st, stream, s.traceInfo(st, stream)) ——> s.processUnaryRPC(t, stream, srv, md, trInfo)

processUnaryRPC 方法中有一段 server 发送响应数据的代码。其实也就是这一行：

```go
if err := s.sendResponse(t, stream, reply, cp, opts, comp); err != nil {
```

其实也能猜到，发送数据给 client 之前肯定要编码。果然调用了 encode 方法

```go
func (s *Server) sendResponse(t transport.ServerTransport, stream *transport.Stream, msg interface{}, cp Compressor, opts *transport.Options, comp encoding.Compressor) error {
    data, err := encode(s.getCodec(stream.ContentSubtype()), msg)
    if err != nil {
        grpclog.Errorln("grpc: server failed to encode response: ", err)
        return err
    }
    ...
}
```

来看一下 encode

```go
func encode(c baseCodec, msg interface{}) ([]byte, error) {
    if msg == nil { // NOTE: typed nils will not be caught by this check
        return nil, nil
    }
    b, err := c.Marshal(msg)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "grpc: error while marshaling: %v", err.Error())
    }
    if uint(len(b)) > math.MaxUint32 {
        return nil, status.Errorf(codes.ResourceExhausted, "grpc: message too large (%d bytes)", len(b))
    }
    return b, nil
}
```

它调用了 c.Marshal 方法， Marshal 方法其实是 baseCodec 定义的一个通用抽象方法

```go
type baseCodec interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
}
```

proto 实现了 baseCodec，前面说到了通过 s.getCodec(stream.ContentSubtype(),msg) 获取到的其实是 contentType 里面设置的协议名称，不设置的话默认取 proto 的编码器。所以最终是调用了 proto 包下的 Marshal 方法，如下：

```go
func (codec) Marshal(v interface{}) ([]byte, error) {
    if pm, ok := v.(proto.Marshaler); ok {
        // object can marshal itself, no need for buffer
        return pm.Marshal()
    }
    cb := protoBufferPool.Get().(*cachedProtoBuffer)
    out, err := marshal(v, cb)
    // put back buffer and lose the ref to the slice
    cb.SetBuf(nil)
    protoBufferPool.Put(cb)
    return out, err
}
```

至此，grpc 的整个编解码的流程我们就已经剖析完了















