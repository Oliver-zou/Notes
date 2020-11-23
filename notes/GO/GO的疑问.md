<!-- GFM-TOC -->

* [一 、赋值](#一-赋值)

    

    <!-- GFM-TOC -->

# 一 、赋值

问题1.

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