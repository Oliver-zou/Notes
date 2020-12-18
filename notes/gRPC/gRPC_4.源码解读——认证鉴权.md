<!-- GFM-TOC -->

* [一 、概述](#一-概述)
  - [1. 单体模式下的认证鉴权](#1. 单体模式下的认证鉴权)
  - [2. 微服务模式下的认证鉴权](#2. 微服务模式下的认证鉴权)
  - [3. grpc 认证鉴权](#3. grpc 认证鉴权)
* [二、代码实现](#二-代码实现)
  - [2.1 使用证书进行 TLS 通信认证](#2.1 使用证书进行 TLS 通信认证)
  - [2.2 oauth2](#2.2 oauth2)

<!-- GFM-TOC -->

# 一、概述

### **1. 单体模式下的认证鉴权**

在单体模式下，整个应用是一个进程，应用一般只需要一个统一的安全认证模块来实现用户认证鉴权。例如用户登陆时，安全模块验证用户名和密码的合法性。假如合法，为用户生成一个唯一的 Session。将 SessionId 返回给客户端，客户端一般将 SessionId 以 Cookie 的形式记录下来，并在后续请求中传递 Cookie 给服务端来验证身份。为了避免 Session Id被第三者截取和盗用，客户端和应用之前应使用 TLS 加密通信，session 也会设置有过期时间。

客户端访问服务端时，服务端一般会用一个拦截器拦截请求，取出 session id，假如 id 合法，则可判断客户端登陆。然后查询用户的权限表，判断用户是否具有执行某次操作的权限。

### **2. 微服务模式下的认证鉴权**

在微服务模式下，一个整体的应用可能被拆分为多个微服务，之前只有一个服务端，现在会存在多个服务端。对于客户端的单个请求，为保证安全，需要跟每个微服务都要重复上面的过程。这种模式每个微服务都要去实现相同的校验逻辑，肯定是非常冗余的。

**用户身份认证**

为了避免每个服务端都进行重复认证，采用一个服务进行统一认证。所以考虑一个单点登录的方案，用户只需要登录一次，就可以访问所有微服务。一般在 api 的 gateway 层提供对外服务的入口，所以可以在 api gateway 层提供统一的用户认证。

**用户状态保持**

由于 http 是一个无状态的协议，前面说到了单体模式下通过 cookie 保存用户状态， cookie 一般存储于浏览器中，用来保存用户的信息。但是 cookie 是有状态的。客户端和服务端在一次会话期间都需要维护 cookie 或者 sessionId，在微服务环境下，我们期望服务的认证是无状态的。所以我们一般采用 token 认证的方式，而非 cookie。

token 由服务端用自己的密钥加密生成，在客户端登录或者完成信息校验时返回给客户端，客户端认证成功后每次向服务端发送请求带上 token，服务端根据密钥进行解密，从而校验 token 的合法，假如合法则认证通过。token 这种方式的校验不需要服务端保存会话状态。方便服务扩展

**标准的http协议是无状态的，无连接的**

无连接：限制每次连接只处理一个请求。服务器处理完客户的请求，并收到客户的应答后，即断开连接

无状态：HTTP 协议自 身不对请求和响应之间的通信状态进行保存。也就是说在 HTTP 这个 级别，协议对于发送过的请求或响应都不做持久化处理。主要是为了让 HTTP 协议尽可能简单，使得它能够处理大量事务。HTTP/1.1 引入 Cookie 来保存状态信息。

1. **服务要设计为无状态的，这主要是从可伸缩性来考虑的。**

2. 如果server是无状态的，那么对于客户端来说，就可以将请求发送到任意一台server上，然后就可以通过**负载均衡**等手段，实现**水平扩展**。

3. 如果server是有状态的，那么就无法很容易地实现了，因为客户端需要始终把请求发到同一台server才行，所谓*“**session迁移”***等方案，也就是为了解决这个问题 

### **3. grpc 认证鉴权**

grpc-go 官方对于认证鉴权的介绍如下：https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-auth-support.md

通过官方介绍可知， grpc-go 认证鉴权是通过 tls + oauth2 实现的。这里不对 tls 和 oauth2 进行详细介绍，假如有不清楚的可以参考阮一峰老师的教程，介绍得比较清楚[tls](http://www.ruanyifeng.com/blog/2014/02/ssl_tls.html ) , [oauth2](http://www.ruanyifeng.com/blog/2019/04/oauth_design.html)

下面我们就来具体看看 grpc-go 是如何实现认证鉴权的)

grpc-go 官方 doc 说了这里关于 auth 的部分有 demo 放在 examples 目录下的 features 目录下。但是 demo 没有包括证书生成的步骤，这里我们自建一个 demo，从生成证书开始一步步进行 grpc 的认证讲解。

# 二、代码实现

生成私钥

```shell
openssl ecparam -genkey -name secp384r1 -out server.key
```

使用私钥生成证书

```shell
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

填写信息（注意 Common Name 要填写服务名）

```shell
Country Name (2 letter code) []:
State or Province Name (full name) []:
Locality Name (eg, city) []:
Organization Name (eg, company) []:
Organizational Unit Name (eg, section) []:
Common Name (eg, fully qualified host name) []:
helloauthEmail Address []:
```

#### 2.1 使用证书进行 TLS 通信认证

之前的 helloworld demo 中，client 在创建 DialContext 指定非安全模式通信，如下：

```
 复制代码    conn, err := grpc.Dial(address, grpc.WithInsecure())
```

这种模式下，client 和 server 都不会进行通信认证，其实是不安全的。下面来看看安全模式下应该如何通信：

**server**

```go
// Package main implements a server for Greeter service.
package main

import (
   "context"
   "log"
   "net"

   "google.golang.org/grpc"
   "google.golang.org/grpc/credentials"

   pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
   port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct{

}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
   log.Printf("Received: %v", in.Name)
   return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
   c, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")
       if err != nil {
           log.Fatalf("credentials.NewServerTLSFromFile err: %v", err)
       }

   lis, err := net.Listen("tcp", port)
   if err != nil {
      log.Fatalf("failed to listen: %v", err)
   }
   s := grpc.NewServer(grpc.Creds(c))
   pb.RegisterGreeterServer(s, &server{})
   if err := s.Serve(lis); err != nil {
      log.Fatalf("failed to serve: %v", err)
   }
}
```

**client**

```go
// Package main implements a client for Greeter service.
package main

import (
   "context"
   "log"
   "os"
   "time"

   "google.golang.org/grpc"
   "google.golang.org/grpc/credentials"

   pb "helloauth/helloworld"
)

const (
   address     = "localhost:50051"
   defaultName = "world"
)

func main() {
   cred, err := credentials.NewClientTLSFromFile("../keys/server.pem", "helloauth")
       if err != nil {
           log.Fatalf("credentials.NewClientTLSFromFile err: %v", err)
       }

   // Set up a connection to the server.
   // 指定非安全模式通信
   conn, err := grpc.Dial(address, grpc.WithTransportCredentials(cred))
   if err != nil {
      log.Fatalf("did not connect: %v", err)
   }
   defer conn.Close()
   c := pb.NewGreeterClient(conn)

   // Contact the server and print out its response.
   name := defaultName
   if len(os.Args) > 1 {
      name = os.Args[1]
   }
   ctx, cancel := context.WithTimeout(context.Background(), time.Second)
   defer cancel()
   r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
   if err != nil {
      log.Fatalf("could not greet: %v", err)
   }
   log.Printf("Greeting: %s", r.Message)
}
```

##### grpc 认证鉴权源码解读

**server**

先来看 server 端，server 端根据 server 的公钥和私钥生成了一个 TransportCredentials ，如下：

```go
c, err := credentials.NewServerTLSFromFile("../keys/server.pem", "../keys/server.key")

// NewServerTLSFromFile constructs TLS credentials from the input certificate file and key
// file for server.
func NewServerTLSFromFile(certFile, keyFile string) (TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}}), nil
}
```

看一下 NewTLS 这个方法，他其实就返回了一个 tlsCreds 的结构体，这个结构体实现了 TransportCredentials 这个接口，包括 ClientHandshake 和 ServerHandshake 。

```go
// NewTLS uses c to construct a TransportCredentials based on TLS.
func NewTLS(c *tls.Config) TransportCredentials {
	tc := &tlsCreds{cloneTLSConfig(c)}
	tc.config.NextProtos = appendH2ToNextProtos(tc.config.NextProtos)
	return tc
}
```

来看一下服务端握手的方法 ServerHandshake，可以发现其底层还是调用 go 的 tls 包去实现 tls 认证鉴权。

```go
func (c *tlsCreds) ServerHandshake(rawConn net.Conn) (net.Conn, AuthInfo, error) {
	conn := tls.Server(rawConn, c.config)
	if err := conn.Handshake(); err != nil {
		return nil, nil, err
	}
	return internal.WrapSyscallConn(rawConn, conn), TLSInfo{conn.ConnectionState()}, nil
}
```

**client**

和 server 端类似，client 端也是通过公钥和服务名先创建一个 TransportCredentials

```go
cred, err := credentials.NewClientTLSFromFile("../keys/server.pem", "helloauth")
```

看一下 NewClientTLSFromFile 这个方法，发现它也是调用了相同的 NewTLS 方法返回了一个 tlsCreds 结构体

```go
func NewTLS(c *tls.Config) TransportCredentials {
    tc := &tlsCreds{cloneTLSConfig(c)}
    tc.config.NextProtos = appendH2ToNextProtos(tc.config.NextProtos)
    return tc
}
```

接下来在创建客户端连接时，将 tlsCreds 这个结构体传了进去。

```go
conn, err := grpc.Dial(address, grpc.WithTransportCredentials(cred))
```

Dial —— > DialContext 方法中有这么一段代码，将我们传入的 serverName 也就是 “helloauth” 赋值给了 clientConn 的 authority 这个字段。

```go
    creds := cc.dopts.copts.TransportCredentials
    if creds != nil && creds.Info().ServerName != "" {
        cc.authority = creds.Info().ServerName
    } else if cc.dopts.insecure && cc.dopts.authority != "" {
        cc.authority = cc.dopts.authority
    } else {
        // Use endpoint from "scheme://authority/endpoint" as the default
        // authority for ClientConn.
        cc.authority = cc.parsedTarget.Endpoint
    }
```

##### 认证过程

**client**

那什么时候开始认证呢？先来说说 client。

client 的认证其实是在调用 connect 方法的时候，在之前讲述负载均衡时降到了，在 acBalancerWrapper 里面有一个 UpdateAddresses 方法，调用 ac.connect() ——> ac.resetTransport() ——> ac.tryAllAddrs ——> ac.createTransport ——> transport.NewClientTransport ——> newHTTP2Client 方法时，有这么一段代码：

transportCreds := opts.TransportCredentials perRPCCreds := opts.PerRPCCredentials

```go
    if b := opts.CredsBundle; b != nil {
        if t := b.TransportCredentials(); t != nil {
            transportCreds = t
        }
        if t := b.PerRPCCredentials(); t != nil {
            perRPCCreds = append(perRPCCreds, t)
        }
    }
    if transportCreds != nil {
        scheme = "https"
        conn, authInfo, err = transportCreds.ClientHandshake(connectCtx, addr.Authority, conn)
        if err != nil {
            return nil, connectionErrorf(isTemporary(err), err, "transport: authentication handshake failed: %v", err)
        }
        isSecure = true
    }
```

这里即调用了tlsCreds 的 ClientHandshake 方法进行握手，实现客户端的认证。

**server**

server 的认证其实是在调用 Serve ——> handleRawConn ——> useTransportAuthenticator 方法，调用了 s.opts.creds.ServerHandshake(rawConn) 方法，其底层也是调用 tlsCreds ServerHandshake 方法进行服务端握手。

```go
func (s *Server) useTransportAuthenticator(rawConn net.Conn) (net.Conn, credentials.AuthInfo, error) {
    if s.opts.creds == nil {
        return rawConn, nil, nil
    }
    return s.opts.creds.ServerHandshake(rawConn)
}
```

#### 2.2 oauth2

tls 保证了 client 和 server 通信的安全性，但是无法做到接口级别的权限控制。例如有 A、B、C、D 四个系统，存在下面两个场景： 1、我们希望 A 可以访问 B、C 系统，但是不能访问 D 系统 2、B 系统提供了 b1、b2、b3 三个接口，我们希望 A 系统可以访问 b1、b2 接口，但是不能访问 b3 接口。 此时 tls 认证肯定是无法实现上面两个诉求的，对于这两个场景，grpc 提供了 oauth2 的认证方式。

grpc 官方提供了对 oauth2 认证鉴权的实现 demo，放在 examples 目录的 features 目录的 authentication 目录下。

**server** 端源码实现如下：

```go
func main() {
   flag.Parse()
   fmt.Printf("server starting on port %d...\n", *port)

   cert, err := tls.LoadX509KeyPair(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
   if err != nil {
      log.Fatalf("failed to load key pair: %s", err)
   }
   opts := []grpc.ServerOption{
      // The following grpc.ServerOption adds an interceptor for all unary
      // RPCs. To configure an interceptor for streaming RPCs, see:
      // https://godoc.org/google.golang.org/grpc#StreamInterceptor
      grpc.UnaryInterceptor(ensureValidToken),
      // Enable TLS for all incoming connections.
      grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
   }
   s := grpc.NewServer(opts...)
   pb.RegisterEchoServer(s, &ecServer{})
   lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
   if err != nil {
      log.Fatalf("failed to listen: %v", err)
   }
   if err := s.Serve(lis); err != nil {
      log.Fatalf("failed to serve: %v", err)
   }
}
```

server 端先调用了 tls 包下的 LoadX509KeyPair，通过 server 的公钥和私钥生成了一个 Certificate 结构体来保存证书信息。然后注册了一个校验 token 的方法到拦截器中，并将证书信息设置到 serverOption 中，构造 server 的时候层层透传进去，最终会被设置到 Server 里面 ServerOptions 结构中的 credentials.TransportCredentials 和 UnaryServerInterceptor 中。

来看看这两个结构什么时候会被调用，先梳理调用链路，在 s.Serve ——> s.handleRawConn ——> s.serveStreams ——> s.handleStream ——> s.processUnaryRPC 方法中有一行

```go
reply, appErr := md.Handler(srv.server, ctx, df, s.opts.unaryInt)
```

可以看到调用了 md.Handler 方法，将 s.opts.unaryInt 这个结构传入了进去。s.opts.unaryInt 就是我们之前注册的 UnaryServerInterceptor 拦截器。md 是一个 MethodDesc 这个结构，包括了 MethodName 和 Handler

```go
type MethodDesc struct {
    MethodName string
    Handler    methodHandler
}
```

这里会取出我们之前注册进去的结构，还记得我们介绍 helloworld 时 RegisterService 吗？至于如何取出 MethodName，源码中的设计非常复杂，经过了层层包装，这里不是本节重点就不赘述了。

```go
func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
    s.RegisterService(&_Greeter_serviceDesc, srv)
}
var _Greeter_serviceDesc = grpc.ServiceDesc{
    ServiceName: "helloworld.Greeter",
    HandlerType: (*GreeterServer)(nil),
    Methods: []grpc.MethodDesc{
        {
            MethodName: "SayHello",
            Handler:    _Greeter_SayHello_Handler,
        },
    },
    Streams:  []grpc.StreamDesc{},
    Metadata: "helloworld.proto",
}
```

看到 md.Handler 其实是 _Greeter_SayHello_Handler 这个结构，它也是在 pb 文件中生成的。

```go
func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
    in := new(HelloRequest)
    if err := dec(in); err != nil {
        return nil, err
    }
    if interceptor == nil {
        return srv.(GreeterServer).SayHello(ctx, in)
    }
    info := &grpc.UnaryServerInfo{
        Server:     srv,
        FullMethod: "/helloworld.Greeter/SayHello",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
    }
    return interceptor(ctx, in, info, handler)
}
```

这里调用了我们传入的 interceptor 方法。回到我们的调用：

```go
reply, appErr := md.Handler(srv.server, ctx, df, s.opts.unaryInt)
```

以其实是调用了 s.opts.unaryInt 这个拦截器。这个拦截器是我们之前在 创建 server 的时候赋值的。

```go
    opts := []grpc.ServerOption{
        // The following grpc.ServerOption adds an interceptor for all unary
        // RPCs. To configure an interceptor for streaming RPCs, see:
        // https://godoc.org/google.golang.org/grpc#StreamInterceptor
        grpc.UnaryInterceptor(ensureValidToken),
        // Enable TLS for all incoming connections.
        grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
    }
    s := grpc.NewServer(opts...)
```

看 grpc.UnaryInterceptor 这个方法，其实是将 ensureValidToken 这个函数赋值给了 s.opts.unaryInt

```go
    func UnaryInterceptor(i UnaryServerInterceptor) ServerOption {
        return newFuncServerOption(func(o *serverOptions) {
            if o.unaryInt != nil {
                panic("The unary server interceptor was already set and may not be reset.")
            }
            o.unaryInt = i
        })
    }
```

所以之前我们执行的这一行

```go
    return interceptor(ctx, in, info, handler)
```

其实是执行了 ensureValidToken 这个函数，这个函数就是我们在 server 端定义的 token 校验的函数。先取出我们传入的 metadata 数据，然后校验 token

```go
    func ensureValidToken(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, errMissingMetadata
        }
        // The keys within metadata.MD are normalized to lowercase.
        // See: https://godoc.org/google.golang.org/grpc/metadata#New
        if !valid(md["authorization"]) {
            return nil, errInvalidToken
        }
        // Continue execution of handler after ensuring a valid token.
        return handler(ctx, req)
}
```

校验完 token 后，最终执行了 handler(ctx, req)

```go
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
    }
    return interceptor(ctx, in, info, handler)
```

可以看到最终其实执行了 GreeterServer 的 SayHello 这个函数，也就是我们在 main 函数中定义的，这个函数就是我们在 server 端定义的提供 SayHello 给客户端回消息的函数。

```go
// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
    log.Printf("Received: %v", in.Name)
    return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}
```

这里还可以额外说一下，md.Handler 执行完之后，其实 reply 就是 SayHello 的回包。

```go
    reply, appErr := md.Handler(srv.server, ctx, df, s.opts.unaryInt)
```

获取到回包之后 server 执行了 sendResponse 方法，将回包发送给 client，这个方法我们之前已经剖析过了，最终会调用 http2Server 的 Write 方法。

```go
if err := s.sendResponse(t, stream, reply, cp, opts, comp); err != nil {
```

看到这里，server 端对 token 的校验在哪里执行的我们已经清楚了。

**client**

先看 main 函数

```go
	// Set up the credentials for the connection.
	perRPC := oauth.NewOauthAccess(fetchToken())
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "x.test.example.com")
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}
	opts := []grpc.DialOption{
		// In addition to the following grpc.DialOption, callers may also use
		// the grpc.CallOption grpc.PerRPCCredentials with the RPC invocation
		// itself.
		// See: https://godoc.org/google.golang.org/grpc#PerRPCCredentials
		grpc.WithPerRPCCredentials(perRPC),
		// oauth.NewOauthAccess requires the configuration of transport
		// credentials.
		grpc.WithTransportCredentials(creds),
	}

	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	rgc := ecpb.NewEchoClient(conn)

	callUnaryEcho(rgc, "hello world")
```

可以看到 client 首先通过 NewOauthAccess 方法生成了包含 token 信息的 PerRPCCredentials 结构

```go
// NewOauthAccess constructs the PerRPCCredentials using a given token.
func NewOauthAccess(token *oauth2.Token) credentials.PerRPCCredentials {
   return oauthAccess{token: *token}
}
```

然后再将 PerRPCCredentials 通过 grpc.WithPerRPCCredentials(perRPC) 添加到了到了 client 的 DialOptions 中的 transport.ConnectOptions 结构中的 [] credentials.PerRPCCredentials 结构中。

那么这个结构什么时候被使用呢，我们来看看。先梳理下调用链 ，在 client 发起rpc调用的 Invoke ——> invoke ——> newClientStream ——> cs.newAttemptLocked ——> cs.cc.getTransport ——> pick ——> acw.getAddrConn().getReadyTransport() ——> ac.connect() ——> ac.resetTransport() ——> ac.tryAllAddrs ——> ac.createTransport ——> transport.NewClientTransport ——> newHTTP2Client 这个方法里面，有这么一段代码，先取出 []credentials.PerRPCCredentials 中的所有 PerRPCCredentials 添加到 perRPCCreds 中。

```
    transportCreds := opts.TransportCredentials
    perRPCCreds := opts.PerRPCCredentials
    if b := opts.CredsBundle; b != nil {
        if t := b.TransportCredentials(); t != nil {
            transportCreds = t
        }
        if t := b.PerRPCCredentials(); t != nil {
            perRPCCreds = append(perRPCCreds, t)
        }
    }
```

然后再将 perRPCCreds 赋值给 http2Client 的 perRPCCreds 属性

```go
t := &http2Client{
    ...
    perRPCCreds:           perRPCCreds,
    ...
}
```

那么 perRPCCreds 属性什么时候被用呢？来继续跟踪，newClientStream 方法中有一段代码

```go
    op := func(a *csAttempt) error { return a.newStream() }
    if err := cs.withRetry(op, func() { cs.bufferForRetryLocked(0, op) }); err != nil {
        cs.finish(err)
        return nil, err
    }
```

这里调用了 csAttempt 的 newStream ——> a.t.NewStream (http2Client 的 NewStream) ——> createHeaderFields ——> getTrAuthData 方法

```go
func (t *http2Client) getTrAuthData(ctx context.Context, audience string) (map[string]string, error) {
    if len(t.perRPCCreds) == 0 {
        return nil, nil
    }
    authData := map[string]string{}
    for _, c := range t.perRPCCreds {
        data, err := c.GetRequestMetadata(ctx, audience)
        if err != nil {
            if _, ok := status.FromError(err); ok {
                return nil, err
            }
            return nil, status.Errorf(codes.Unauthenticated, "transport: %v", err)
        }
        for k, v := range data {
            // Capital header names are illegal in HTTP/2.
            k = strings.ToLower(k)
            authData[k] = v
        }
    }
    return authData, nil
}
```

通过调用 GetRequestMetadata 取出 token 信息，这里会调用 oauth 的 GetRequestMetadata 方法 ，按照指定格式拼装成一个 map[string]string{} 的形式

```go
func (s *serviceAccount) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if !s.t.Valid() {
        var err error
        s.t, err = s.config.TokenSource(ctx).Token()
        if err != nil {
            return nil, err
        }
    }
    return map[string]string{
        "authorization": s.t.Type() + " " + s.t.AccessToken,
    }, nil
}
```

然后将以 map[string]string{} 的形式组装成一个 string map 返回，如下：

```go
   for k, v := range authData {
        headerFields = append(headerFields, hpack.HeaderField{Name: k, Value: encodeMetadataHeader(k, v)})
    }
```

返回的 map 会被遍历每个 key，并设置到 headerFields 中，以 http 头部的形式发送出去。数据最终会被 metadata.FromIncomingContext(ctx) 获取到，然后被取出 map 数据。

至此，client 和 server 的数据流转过程被打通