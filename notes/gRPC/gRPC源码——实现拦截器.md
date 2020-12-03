<!-- GFM-TOC -->

* [一 、概述](#一-概述)
* [二、定义](#二-定义)

- [三、参考gPRC-go的example](#三-参考gPRC-go的example)
- [四、最终实现](#四-最终实现)
- [参考](#参考)

<!-- GFM-TOC -->

# 一、概述

拦截器，在AOP（Aspect-Oriented Programming）中用于在某个方法或字段被访问之前，进行拦截然后在之前或之后加入某些操作。拦截是AOP的一种实现策略。通俗点说，就是在执行一段代码（二、中的handler）之前或者之后，去执行另外一段代码。

接下来用Go实现一个拦截器，假设有一个方法 handler(ctx context.Context) ，给这个方法赋予一个能力：允许在这个方法执行之前能够打印一行日志。

# 二、定义

**2.1 结构体**

定义一个结构 interceptor 这个结构包含两个参数，一个 context 和 一个 handler

```go
// 将 handler 单独定义成一种类型
type handler func(ctx context.Context)

type interceptor func(ctx context.Context, h handler)
```

**2.2 申明赋值+编写main函数**

```go
func main() {
    var ctx context.Context

    var ceps []interceptor
  	// 申明赋值
		// 为了实现目标，对 handler 的每个操作，都需要先经过 interceptor ，
  	// 于是申明两个 interceptor 和 handler 的变量并赋值
    var h = func(ctx context.Context) {
        fmt.Println("do something ...")
    }

    var inter1 = func(ctx context.Context, h handler) {
        fmt.Println("interceptor1")
        h(ctx)
    }
    var inter2 = func(ctx context.Context, h handler) {
        fmt.Println("interceptor2")
        h(ctx)
    }

    ceps = append(ceps, inter1, inter2)

    for _ , cep := range ceps {
        cep(ctx, h)
    }
}
/////////////////////////////////////////////////////
interceptor1
do something ...
interceptor2
do something ...
```

handler执行了两次，与预期效果不同，希望无论打印多少次内容，应该保证handler只执行一次（也就是拦截多次，handler只有一次）。

# 三、参考gPRC-go的example

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

# 四、最终实现

**4.1 结构体**

将原来的 handler 升级一下，成为 Invoker , 重新定义一个 handler ，用于在 Invoker 执行之前处理某些事情。interceptor 也需要更改一下，需要传入 invoker 和 handler

```go
type invoker func(ctx context.Context, interceptors []interceptor2 , h handler) error
type handler func(ctx context.Context)

type interceptor2 func(ctx context.Context, h handler, ivk invoker) error
```

**4.2 串联结构体**

```go
func getInvoker(ctx context.Context, interceptors []interceptor2 , cur int, ivk invoker) invoker{
     if cur == len(interceptors) - 1 {
        return ivk
    }
     return func(ctx context.Context, interceptors []interceptor2 , h handler) error{
        return     interceptors[cur+1](ctx, h, getInvoker(ctx,interceptors, cur+1, ivk))
    }
}
```

**4.3 返回第一个 interceptor 作为入口**

```go
func getChainInterceptor(ctx context.Context, interceptors []interceptor2 , ivk invoker) interceptor2 {
        if len(interceptors) == 0 {
            return nil
        }
        if len(interceptors) == 1 {
            return interceptors[0]
        }
        return func(ctx context.Context, h handler, ivk invoker) error {
            return interceptors[0](ctx, h, getInvoker(ctx, interceptors, 0, ivk))
        }
    }
```

**4.4 最终实现**

```go
type interceptor2 func(ctx context.Context, h handler, ivk invoker) error

type handler func(ctx context.Context)

type invoker func(ctx context.Context, interceptors []interceptor2 , h handler) error

func main() {

    var ctx context.Context
    var ceps []interceptor2
    var h = func(ctx context.Context) {
        fmt.Println("do something")
    }
    var inter1 = func(ctx context.Context, h handler, ivk invoker) error{
        h(ctx)
        return ivk(ctx,ceps,h)
    }
    var inter2 = func(ctx context.Context, h handler, ivk invoker) error{
        h(ctx)
        return ivk(ctx,ceps,h)
    }

    var inter3 = func(ctx context.Context, h handler, ivk invoker) error{
        h(ctx)
        return     ivk(ctx,ceps,h)
    }

    ceps = append(ceps, inter1, inter2, inter3)
    var ivk = func(ctx context.Context, interceptors []interceptor2 , h handler) error {
        fmt.Println("invoker start")
        return nil
    }

    cep := getChainInterceptor(ctx, ceps,ivk)
    cep(ctx, h,ivk)

}

func getChainInterceptor(ctx context.Context, interceptors []interceptor2 , ivk invoker) interceptor2 {
    if len(interceptors) == 0 {
        return nil
    }
    if len(interceptors) == 1 {
        return interceptors[0]
    }
    return func(ctx context.Context, h handler, ivk invoker) error {
        return interceptors[0](ctx, h, getInvoker(ctx, interceptors, 0, ivk))
    }

}


func getInvoker(ctx context.Context, interceptors []interceptor2 , cur int, ivk invoker) invoker{
     if cur == len(interceptors) - 1 {
        return ivk
    }
     return func(ctx context.Context, interceptors []interceptor2 , h handler) error{
        return     interceptors[cur+1](ctx, h, getInvoker(ctx,interceptors, cur+1, ivk))
    }
}
////////////////////////////////////////////////////////////
do something
do something
do something
invoker start
```

可以看到每次 Invoker 执行前我们都调用了 handler，但是 Invoker 只被调用了一次，完美地实现了我们的诉求，一个简化版的拦截器诞生了。