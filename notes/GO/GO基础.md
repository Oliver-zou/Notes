<!-- GFM-TOC -->

* [一 、赋值](#一-赋值)

* [二、传递or引用](#二-传递or引用)
  - [1.值类型](#1-值类型)
  - [2.引用类型](#2-引用类型)
* 三、[用法](#三-用法)

  - [1.Golang 中函数作为值与类型](1-Golang 中函数作为值与类型)
  - [2.for的用法](1-for的用法)

* [记录点](记录点)

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

值传递：在调用函数时将实际参数复制一份传递到函数中，这样在函数中如果对参数进行修改，将不会影响到实际参数。

引用传递：调用函数时将实际参数的地址传递到函数中，那么在函数中对参数所进行的修改，将影响到实际参数。

队列：先进先出
栈：先进后出，在程序调用的时候从栈空间去分配（值数据类型通常在此）
堆：在程序调用的时候从系统的内存区分配（引用数据类型在堆区分配空间）

但是！！关于参数传递，Golang文档中有这么一句:

> after they are evaluated, the parameters of the call are passed by value to the
> function and the called function begins execution.
>
> 函数调用参数**均为值传递**，不是指针传递或引用传递。经测试引申出来，当参数变量为指针或隐式指针类型，参数传递方式也是传值（指针本身的copy）

也就是说，**Go的参数传递都是值传递**，但像指针、`map`、`chan`等类型的变量在参数传递中可以被修改呢？下面进行具体说明

## 1.值类型

值类型：基本数据类型int, float,bool, string以及数组和struct

```go
值类型：变量直接存储值，内容通常在栈中分配
 var i = 5       i -----> 5
```

**1.1基本数据类型**

对于基本数据类型，形参是实参的拷贝

```go
func modify(inside int) {
   inside = 2
   fmt.Printf("inside value = %v\n", inside)
   fmt.Printf("inside address = %p\n", &inside)
}

func main() {
   outside := 1
   modify(outside)
   fmt.Printf("outside value = %v\n", outside)
   fmt.Printf("outside address = %p\n", &outside)
}
/////////////////////////////////////////
inside value = 2
inside address = 0xc00000a0f0
outside value = 1
outside address = 0xc00000a0d8
```

实参`outside`的地址与形参`inside`的地址不同，因此`outside`的值没有改变。

**1.2数组类型**

与基本数据类型同理，将数据拷贝传入

```go
func modify(inside [3]int) {
   inside[2] = 4
   fmt.Printf("inside value = %v\n", inside)
   fmt.Printf("inside address = %p\n", &inside)
}

func main() {
   outside := [3]int{1, 2, 3}
   modify(outside)
   fmt.Printf("outside value = %v\n", outside)
   fmt.Printf("outside address = %p\n", &outside)
}
////////////////////////////////////
inside value = [1 2 4]
inside address = 0xc00009e160
outside value = [1 2 3]
outside address = 0xc00009e140
```

实参数组`outside`的起始地址与形参数组`inside`的起始地址不同，即`inside`是`outside`的拷贝，修改`inside`的元素不会改变`outside`对应的元素。

**1.3结构体**

```go
type IdCard struct {
   Id int32
}
type Person struct {
   Card *IdCard
   Age int32
}

func modify(inside Person) {
   inside.Card.Id = 456
   inside.Age = 20
   fmt.Printf("inside Card value = %v[address=%p]\n", inside.Card, inside.Card)
   fmt.Printf("inside Age value = %v\n", inside.Age)
   fmt.Printf("inside address = %p\n", &inside)
}

func main() {
   outside := Person {
      Card: &IdCard{
         Id: 123,
      },
      Age: 18,
   }
   modify(outside)
   fmt.Printf("outside Card value = %v[address=%p]\n", outside.Card, outside.Card)
   fmt.Printf("outside Age value = %v\n", outside.Age)
   fmt.Printf("outside address = %p\n", &outside)
}

//////////////////////////////////////
inside Card value = &{456}[address=0xc00000a0d8]
inside Age value = 20
inside address = 0xc00003c200
outside Card value = &{456}[address=0xc00000a0d8]
outside Age value = 18
outside address = 0xc00003c1f0
```

可以看出`struct`类型也是值传递（`inside`是将`outside`值拷贝一份，保存在新的内存地址）。在拷贝过程中，指针`Card`只是拷贝了指针的值而没有分配新的内存空间，即发生的是浅拷贝，因此`inside.Card`和`outside.Card`指向的是同一块内存空间，所以修改`inside.Card`的值同时也修改了。

## 2.引用类型

引用类型：指针，slice，map，chan等都是引用类型

```go
引用类型：变量存储的是一个地址，这个地址存储最终的值，内容通常在堆上分配，通过GC回收
ref r ------> 内存地址 -----> 值
```

**2.1指针**

**值传递**：形参时实参的拷贝，改变函数形参并不影响函数外部的实参，这是最常用的一种传递方式，也是最简单的一种传递方式。只需要传递参数，返回值是return考虑的；使用值传递这种方式，调用函数不对实参进行操作，也就是说，即使形参的值发生改变，实参的值也完全不受影响。

**指针传递**其实是值传递的一种，它传递的是地址。值传递过程中，被调函数的形参作为被调函数的局部变量来处理，即在函数的栈中有开辟了内存空间来存放主调函数放进来实参的值，从而成为一个副本。因为指针传递的是外部参数的地址，当调用函数的形参发生改变时，自然外部实参也发生改变。

**引用传递**：被调函数的形参虽然也作为局部变量在栈中开辟了内存空间，但在栈中放的是由主调函数放进来的实参变量的地址。被调函数对形参的任何操作都被间接寻址，即通过栈中的存放的地址访问主调函数中的中的实参变量（相当于一个人有两个名字），因此形参在任意改动都直接影响到实参。

```go
func modify(inside *int) {
   *inside = 2
   fmt.Printf("inside value = %v\n", inside)
   fmt.Printf("inside pointed value = %v\n", *inside)
   fmt.Printf("inside address = %p\n", &inside)
}

func main() {
   num := 1
   outside := &num
   modify(outside)
   fmt.Printf("outside value = %v\n", outside)
   fmt.Printf("outside pointed value = %v\n", *outside)
   fmt.Printf("outside address = %p\n", &outside)
}
///////////////////////////////////
inside value = 0xc00000a0d8
inside pointed value = 2
inside address = 0xc000006030
outside value = 0xc00000a0d8
outside pointed value = 2
outside address = 0xc000006028
```

`inside`是`outside`的拷贝（`inside`和`outside`指向的是同一块内存空间），因此修改`inside`指向的值，`outside`指向的值也会一起改变。

**2.2切片**

切片和数组是不一样的，数组是值类型，切片是引用类型。`slice`底层代码：

```go
// runtime/slice.go
type slice struct {
    array unsafe.Pointer // 指针
    len   int            // 长度 
    cap   int            // 容量
}
// slice就是个结构体，array是底层数组的地址
```

```go
func modify(inside []int) {
	inside[2] = 4
	fmt.Printf("inside value = %v\n", inside)
	fmt.Printf("inside address = %p\n", &inside)
	fmt.Printf("inside first element address = %p\n", &inside[0])
}

func main() {
	outside := []int{1, 2, 3}
	modify(outside)
	fmt.Printf("outside value = %v\n", outside)
	fmt.Printf("outside address = %p\n", &outside)
	fmt.Printf("outside first element address = %p\n", &outside[0])
}
//////////////////////////////////////////////////////
inside value = [1 2 4]
inside address = 0xc0000044c0
inside first element address = 0xc0000123a0
outside value = [1 2 4]
outside address = 0xc0000044a0
outside first element address = 0xc0000123a0
```

可以看到`inside`的地址和`outside`不同，而`inside`第一个元素的地址和`outside`第一个元素的地址相同，这是因为`inside`是`outside`的拷贝，即`inside`中`array`指针的值和`outside`相同。虽然修改`inside`的元素后`outside`的值也被修改了，但是实际上修改的是`inside`和`outside`指向的数组，`inside`和`outside`的值并没有被修改。

当`slice`的长度发生变化时：

```go
func modify(inside []int) {
	inside = append(inside, 4)
	fmt.Printf("inside value = %v\n", inside)
	fmt.Printf("inside address = %p\n", &inside)
	fmt.Printf("inside first element address = %p\n", &inside[0])
}

func main() {
	outside := []int{1, 2, 3}
	modify(outside)
	fmt.Printf("outside value = %v\n", outside)
	fmt.Printf("outside address = %p\n", &outside)
	fmt.Printf("outside first element address = %p\n", &outside[0])
}
//////////////////////////////////////////////////////
inside value = [1 2 3 4]
inside address = 0xc0000044c0
inside first element address = 0xc00000c3f0
outside value = [1 2 3]
outside address = 0xc0000044a0
outside first element address = 0xc0000123a0
```

`outside`并没有随着`inside`的内容发生变化而变化，这是因为append操作无论数组还是切片，都有长度限制。也就是追加切片的时候，如果元素正好在切片的容量范围内，直接在尾部追加一个元素即可。如果超出了最大容量，再追加元素就需要针对底层的数组进行复制和扩容操作了。

因为`slice`在扩容的时候会把原来的底层数组拷贝一份，将`array`指针指向这个新的数组，所以扩容后`inside`元素的地址和`outside`不再相同，修改`inside`的元素也就不会影响`outside`的元素了。

请看以下例子：

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

当函数的参数是切片的时候，到底是传值还是传引用？从`modify`函数中打出的参数的地址，可以看出肯定不是传引用，毕竟引用都是一个地址才对。然而`modify`函数内改变了参数的值，也改变了原始变量slice的值，这个看起来像引用的现象，实际上正是我们前面讨论的切片共享底层数组的实现。

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

**2.3map chan**

概述：

- make map or chan的时候，其实返回的是都是 `hmap` 和 `hchan` 的指针，所以没必要再对它们进行取址。

1. 如果是指向map或者chan的pointer, 这两个本来就是返回了指向对应结构体的指针, 再取个指针用起来麻烦.
2. 至于指针作为值存储, 其实是没啥问题, 只是有的时候使用原生类型或者不含指针的struct效率会高一些.
   map的key, value使用无指针的类型可以减少指针数, 减少了gc扫描时的工作量.
   对于chan的数据类型为非指针的话, 那么在两个协程之间传递数据, go其实是可以直接通过栈对栈的拷贝进行传递的, 小数据的时候, 会提高效率. 如果传递的数据struct很大, 用指针可能会性能更好一些.

```go
func modify(inside map[int64]string) {
	inside[1] = "world"
	fmt.Printf("inside value = %v\n", inside)
	fmt.Printf("inside address = %p\n", &inside)
}

func main() {
	outside := make(map[int64]string, 0)
	outside[1] = "hello"
	modify(outside)
	fmt.Printf("outside value = %v\n", outside)
	fmt.Printf("outside address = %p\n", &outside)
}
///////////////////////////////////
inside value = map[1:world]
inside address = 0xc000006030
outside value = map[1:world]
outside address = 0xc000006028
```

`inside[1]`修改后，`outside[1]`的值也被修改，但inside`和`outside`的地址却不一样，说明还是值传递，而`inside`和`outside`很有可能是指针。下面从源码寻找答案：

```
func makemap(t *maptype, hint int64, h *hmap, bucket unsafe.Pointer) *hmap {
    // ...
}
```

`make(map[int64]string, 0)`返回的确实是`hmap`类型的指针，`func update(inside map[int64]string)`其实就是`func update(inside *hmap)`，这样就可以理解`map`类型的变量作为函数参数传递时为何实参的值也会被修改了。`chan`类型与`map`相似，返回指针`hchan` 。

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











# 记录点

## 1. init()

```go
fun init(){}  // 自动执行
```

注意事项：

\- init函数先于main函数自动执行，不能被其他函数调用；
\- init函数没有输入参数、返回值；
\- 每个包可以有多个init函数；
\- **包的每个源文件也可以有多个init函数**；
\- 同一个包的init执行顺序，golang没有明确定义；
\- 不同包的init函数按照包导入的依赖关系决定执行顺序。

 

## 2.空接口

```go
var XXX interface{} //
```

空接口类型的变量可以保存任何类型的值,空格口类型的变量非常类似于弱类型语言中的变量，未被初始化的interface默认初始值为nil。











## N.记录

1. Context可通过 父协程 取消 所有 子协程（**非 Channel 实现**）

2. `sync.Once` 的 `Do（）` 方法，可保证匿名函数只被执行一次

3. `sync.Pool`生命周期不可控，随时会被GC

4. 创建一个协程，除了 `go func(){}` 还有更简洁的方式：

   ```go
   go agt.EventProcessGroutine()
   ```

5. 主线程 可通过 `var wg sync.WaitGroup()` 管理多个协程的并发问题； 











