<!-- GFM-TOC -->

* [一 、赋值](#一-赋值)

* [二、传递or引用](#二-传递or引用)
- [1.数组与切片](#1-数组与切片)
- [2.map chan](#2-map chan)
  
- 三、[用法](#三-用法)

  - [1.Golang 中函数作为值与类型](1-Golang 中函数作为值与类型)
  - [2.for的用法](1-for的用法)

  ​    

  <!-- GFM-TOC -->

# 一 、赋值

**问题1.基础类型和定义类型**

```go
type A int
var v1 A
var v2 int
v1=v2 //编译错误

type A []int
var v1 A
var v2 []int
v1=v2 //编译通过
```

https://medium.com/golangspec/assignability-in-go-27805bcd5874
这个文章很好的总结了这些case。这个文章已比较好懂。
它有引用golang原文：
https://golang.org/ref/spec#Assignability

A value x is assignable to a variable of type T ("x is assignable to T") if one of the following conditions applies:

- x's type is identical to T.
- x's type V and T have identical underlying types and at least one of V or T is not a defined type.
- T is an interface type and x implements T.
  x is a bidirectional channel value, T is a channel type, x's type V and 
- T have identical element types, and at least one of V or T is not a defined type.
- x is the predeclared identifier nil and T is a pointer, function, slice, map, channel, or interface type.
- x is an untyped constant representable by a value of type T.

type A []int、type A int都是defined type。问题出在underlying type和int是不是defined type。https://golang.org/ref/spec#Types

- int的underlying type就是int。下面是原文。
  Each type T has an underlying type: If T is one of the predeclared boolean, numeric, or string types, or a type literal, the corresponding underlying type is T itself. Otherwise, T's underlying type is the underlying type of the type to which T refers in its type declaration.
- int是defined type。参见bool下面说法。' it is a defined type.'
  https://golang.org/ref/spec#Boolean_types

所以，int是defined type，underlying是int，type A int是defined type，underlying是int，两个defined type，不能直接赋值

[]int不是[defined type](https://golang.org/ref/spec#Type_definitions)。它是[Composite literals](https://golang.org/ref/spec#Composite_literals)

type A struct {
F int
}
也是Composite literals
`type A struct {
F int
}

var v1 A
var v2 struct {
F int
}
v1 = v2`
能成立，[]int能成立，int不行。









# 二、传递or引用

## 1.数组与切片

切片和数组是不一样的，数组是值类型，切片是引用类型。请看下面的例子：

```go
func main(){
	slice := []int{1,2,3}
	fmt.Printf("slice %v, slice address %p\n", slice, &slice)
	slice = changeSlice(slice)
	fmt.Printf("slice %v, slice address %p\n", slice, &slice)
}

func changeSlice(nums []int) []int {
	nums[1] = 111
	return nums
}
//////////////////////////////////////////////////////
slice [1 2 3], slice address 0xc0000044a0
slice [1 111 3], slice address 0xc0000044a0
```

结果确实是外部切片中的值进行了改变，地址没有进行改变，但再下面的代码：

```go
func main(){
	slice := []int{1,2,3}
	fmt.Printf("slice %v, slice address %p\n", slice, &slice)
	slice = changeSlice(slice)
	fmt.Printf("slice %v, slice address %p\n", slice, &slice)
}

func changeSlice(nums []int) []int {
	fmt.Printf("nums: %v, nums addr %p\n", nums, &nums)
	nums[1] = 111
	return nums
}
//////////////////////////////////////////////////////
slice [1 2 3], slice address 0xc0000044a0
nums: [1 2 3], nums addr 0xc000004500
slice [1 111 3], slice address 0xc0000044a0
```

在函数中打印了传入函数的切片地址，发现和外部切片地址并不一样。这就需要引出切片的实现。切片不等同于数组，但其是依赖于数组实现的，切片是一种复合结构，它是由三部分组成的，第一部分是只想底层数组的指针**pt**r，第二部分是切片的大小**len**，最后是切片的容量**cap**。

```go
func main() {
	arr := [5]int{1, 2, 3, 4, 5}
	slice := arr[1:4]
	slice2 := arr[2:5]
		fmt.Printf("slice: %v,slice add %p\n", slice, &slice)
	fmt.Printf("slice2: %v,slice2 add %p\n", slice2, &slice2)
	arr[2] = 11
	fmt.Printf("slice: %v,slice add %p\n", slice, &slice)
	fmt.Printf("slice2: %v,slice2 add %p\n", slice2, &slice2)
}
//////////////////////////////////////////////////////
slice: [2 3 4],slice add 0xc0000044a0
slice2: [3 4 5],slice2 add 0xc0000044c0
slice: [2 11 4],slice add 0xc0000044a0
slice2: [11 4 5],slice2 add 0xc0000044c0
```

有一个5个元素的数组，slice，slice2分别截取了数组的一部分，并且有共同的一部分。可以看出，两个切片公用一个数组，所以一个改变都改变。此外，很明显内存地址不同：从数组中切一块下来形成切片很好理解，有时候我们用make函数创建切片，实际上golang会在底层创建一个匿名的数组。如果从新的slice再切，那么新创建的两个切片都共享这个底层的匿名数组。

若只是引用其他切片的值，而不对他进行改变，那么就需要对切片进行复制：首先肯定要另外开辟一块内存地址，然后进行赋值，内存地址不一样，改变一个的值，不会影响另外一个，go语言中，为了方便复制，也有一个函数就是copy。

append：无论数组还是切片，都有长度限制。也就是追加切片的时候，如果元素正好在切片的容量范围内，直接在尾部追加一个元素即可。如果超出了最大容量，再追加元素就需要针对底层的数组进行复制和扩容操作了。请看以下例子：

```go
func main() {
	arr := [3]int{1, 2, 3}
	slice := arr[1:2]

	fmt.Printf("slice %v, slice addr %p, len %d, cap %d \n", slice, &slice, len(slice), cap(slice))

	slice = append(slice, 222)
	fmt.Printf("slice %v, slice addr %p, len %d, cap %d \n", slice, &slice, len(slice), cap(slice))

	slice = append(slice, 333)
	fmt.Printf("slice %v, slice addr %p, len %d, cap %d \n", slice, &slice, len(slice), cap(slice))
	slice[1] = 111
	fmt.Println(arr, slice)
}
//////////////////////////////////////////////////////
slice [2], slice addr 0xc0000044a0, len 1, cap 2 
slice [2 222], slice addr 0xc0000044a0, len 2, cap 2 
slice [2 222 333], slice addr 0xc0000044a0, len 3, cap 4 
[1 2 222] [2 111 333]
```

当追加超出原本容量时，再改变切片内容后，对原来数组是没有影响的
**slice和array的关系十分密切，通过两者的合理构建，既能实现动态灵活的线性结构，也能提供访问元素的高效性能。当然，这种结构也不是完美无暇，共用底层数组，在部分修改操作的时候，可能带来副作用，同时如果一个很大的数组，那怕只有一个元素被切片应用，那么剩下的数组都不会被垃圾回收，这往往也会带来额外的问题。**



回到最开始的问题，当函数的参数是切片的时候，到底是传值还是传引用？从changeSlice函数中打出的参数s的地址，可以看出肯定不是传引用，毕竟引用都是一个地址才对。然而changeSlice函数内改变了s的值，也改变了原始变量slice的值，这个看起来像引用的现象，实际上正是我们前面讨论的切片共享底层数组的实现。

即切片传递的时候，传的是数组的值，等效于从原始切片中再切了一次。原始切片slice和参数s切片的底层数组是一样的。因此修改函数内的切片，也就修改了数组。

```go
 func main() {
    slice := make([]int, 2, 2)
    for i := 0; i < len(slice); i++ {
        slice[i] = i
    }

    fmt.Printf("slice %v %p \n", slice, &slice)

    ret := changeSlice(slice)
    fmt.Printf("slice %v %p, ret %v \n", slice, &slice, ret)

    ret[1] = -1111

    fmt.Printf("slice %v %p, ret %v \n", slice, &slice, ret)
}

func changeSlice(s []int) []int {
    fmt.Printf("func s %v %p \n", s, &s)
    s[0] = -1
    s = append(s, 3)
    s[1] =  1111
    return s
}
//////////////////////////////////////////////////////
slice [0 1] 0xc0000044a0 
func s [0 1] 0xc000004500 
slice [-1 1] 0xc0000044a0, ret [-1 1111 3] 
slice [-1 1] 0xc0000044a0, ret [-1 -1111 3] 
```

此时append后，明显容量不够用，就会新生成一个底层数组，所以内存地址改变，并且改变新切片对原来的没有任何影响

**通过上面的分析，我们大致可以下结论，slice或者array作为函数参数传递的时候，本质是传值而不是传引用。传值的过程复制一个新的切片，这个切片也指向原始变量的底层数组。（个人感觉称之为传切片可能比传值的表述更准确）。函数中无论是直接修改切片，还是append创建新的切片，都是基于共享切片底层数组的情况作为基础。也就是最外面的原始切片是否改变，取决于函数内的操作和切片本身容量。**

总结：

golang提供了array和slice两种序列结构。其中array是值类型。slice则是复合类型。slice是基于array实现的。slice的第一个内容为指向数组的指针，然后是其长度和容量。通过array的切片可以切出slice，也可以使用make创建slice，此时golang会生成一个匿名的数组。

因为slice依赖其底层的array，修改slice本质是修改array，而array又是有大小限制，当超过slice的容量，即数组越界的时候，需要通过动态规划的方式创建一个新的数组块。把原有的数据复制到新数组，这个新的array则为slice新的底层依赖。

数组还是切片，在函数中传递的不是引用，是另外一种值类型，即通过原始变量进行切片传入。函数内的操作即对切片的修改操作了。当然，如果为了修改原始变量，可以指定参数的类型为指针类型。传递的就是slice的内存地址。函数内的操作都是根据内存地址找到变量本身。

## 2.map chan

make map or chan的时候，其实返回的是都是 `hmap` 和 `hchan` 的指针，所以没必要再对它们进行取址。

1. 如果是指向map或者chan的pointer, 这两个本来就是返回了指向对应结构体的指针, 再取个指针用起来麻烦.
2. 至于指针作为值存储, 其实是没啥问题, 只是有的时候使用原生类型或者不含指针的struct效率会高一些.
   map的key, value使用无指针的类型可以减少指针数, 减少了gc扫描时的工作量.
   对于chan的数据类型为非指针的话, 那么在两个协程之间传递数据, go其实是可以直接通过栈对栈的拷贝进行传递的, 小数据的时候, 会提高效率. 如果传递的数据struct很大, 用指针可能会性能更好一些.







# 三、用法

## 1.Golang 中函数作为值与类型

在 Go 语言中，我们可以把函数作为一种变量，用 type 去定义它，那么这个函数类型就可以作为值传递，甚至可以实现方法，这一特性是在太灵活了，有时候我们甚至可以利用这一特性进行类型转换。作为值传递的条件是类型具有相同的参数以及相同的返回值。

 这一点与python的装饰器的原理类似，区分函数与函数调用的区别

**1.1函数的类型转换**

Go 语言的类型转换基本格式如下：

```go
type_name(expression)
```

下面是例子：

```go
package main	
	
import "fmt"	
	
type CalculateType func(int, int) // 声明了一个函数类型	
	
// 该函数类型实现了一个方法	
func (c *CalculateType) Serve() {	
  fmt.Println("我是一个函数类型")	
}	
	
// 加法函数	
func add(a, b int) {	
  fmt.Println(a + b)	
}	
	
// 乘法函数	
func mul(a, b int) {	
  fmt.Println(a * b)	
}	
	
func main() {	
  a := CalculateType(add) // 将add函数强制转换成CalculateType类型	
  b := CalculateType(mul) // 将mul函数强制转换成CalculateType类型	
  a(2, 3)	
  b(2, 3)	
  a.Serve()	
  b.Serve()	
}	
	
// 5	
// 6	
// 我是一个函数类型	
// 我是一个函数类型
```

如上，声明了一个 CalculateType 函数类型，并实现 Serve() 方法，并将拥有相同参数的 add 和 mul 强制转换成 CalculateType 函数类型，同时这两个函数都拥有了 CalculateType 函数类型的 Serve() 方法。

**1.2函数作参数传递**

```go
package main

import "fmt"

type CalculateType func(a, b int) int // 声明了一个函数类型

// 加法函数
func add(a, b int) int {
  return a + b
}

// 乘法函数
func mul(a, b int) int {
  return a * b
}

func Calculate(a, b int, f CalculateType) int {
  return f(a, b)
}

func main() {
  a, b := 2, 3
  fmt.Println(Calculate(a, b, add))
  fmt.Println(Calculate(a, b, mul))
}
// 5
// 6
复制代码
```

如上例子，Calculate 的 f 参数类型为 CalculateType，add 和 mul 函数具有和 CalculateType 函数类型相同的参数和返回值，因此可以将 add 和 mul 函数作为参数传入 Calculate 函数中。

net/http 包源码例子：

```go
// HandleFunc registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
  DefaultServeMux.HandleFunc(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
  mux.Handle(pattern, HandlerFunc(handler))
}

type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
  f(w, r)
}
```

这段源码的目的是为了将我们的 Handler 强制实现 ServeHTTP() 方法，如下例子：

```go
func sayHi(w http.ResponseWriter, r *http.Request) {
  io.WriteString(w, "hi")
}

func main() {
  http.HandlerFunc("/", sayHi)
  http.ListenAndserve(":8080", nil)
}
```

因为 HandlerFunc 是一个函数类型，而 sayHi 函数拥有和 HandlerFunc 函数类型一样的参数值，因此可以将 sayHi 强制转换成 HandlerFunc，因此 sayHi 也拥有了 ServeHTTP() 方法，也就实现了 Handler 接口，同时，HandlerFunc 的 ServeHTTP 方法执行了它自己本身，也就是 sayHi 函数，这也就可以看出来了，sayHi 就是 Handler 被调用之后的执行结果。

本质上interface{}就是这个道理。

```go
type tt int
//go:noinline
func (c*tt) Test(i int) int {
   return i + 1
}

var rf func(i int) int
func main() {
   var v tt
   rv := reflect.ValueOf(&v)
   method := rv.Method(0)
   rf = method.Interface().(func(i int) int)
   rf(1)  // 调用函数
}

func Benchmark_F1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rf(i)
	}
}
func Benchmark_F2(b *testing.B) {
	var v tt

	f := v.Test
	for i := 0; i < b.N; i++ {
		f(i)
	}
}
```

以上例子还表明，虽然结果相同但调用的函数类型不同，[性能是不一样的](https://github.com/teh-cmc/go-internals/tree/master/chapter2_interfaces)，reflect.object首先从interface{}转换而来后，再是运行期中需要进行的各种耗时操作，导致性能低下:

func (Value) [Interface](https://github.com/golang/go/blob/master/src/reflect/value.go?name=release#1069)

```
func (v Value) Interface() (i interface{})
```

本方法返回v当前持有的值（表示为/保管在interface{}类型），等价于（只是结果等价）：

```
var i interface{} = (v's underlying value)
```

## 2.for的用法

像for一样，if语句可以从简短语句开始，然后在条件之前执行。语句声明的变量仅在范围内，直到if结束。

```go
package main

import (
	"fmt"
	"math"
)

func pow(x, n, lim float64) float64 {
	if v := math.Pow(x, n); v < lim {
		return v
	}
	return lim
}

func main() {
	fmt.Println(
		pow(3, 2, 10),
		pow(3, 3, 20),
	)
}
```



 



 



 











