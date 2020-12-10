<!-- GFM-TOC -->

* [一 、概述](#一-概述)
  - [1. 单体模式下的认证鉴权](#1. 单体模式下的认证鉴权)
  - [2. 微服务模式下的认证鉴权](#2. 微服务模式下的认证鉴权)
  - [3. grpc 认证鉴权](#3. grpc 认证鉴权)
* [二、代码实现](#二-代码实现)

- [三、参考gPRC-go的example](#三-参考gPRC-go的example)
- [参考](#参考)

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

生成完毕后，将证书文件放到 keys 目录下，整个项目目录结构如下：