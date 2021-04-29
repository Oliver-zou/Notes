<!-- GFM-TOC -->

* [一 、基础](#一-基础)

* [二 、装饰器](#一-装饰器)

   


  <!-- GFM-TOC -->

# 一、基础

1. ##### **值传递就是将传递的变量复制了一个，引用传递就是让该元素指向值的地址。**。作为函数参数，可变对象传递的是引用，不可变对象传递的是值[（内容）](https://blog.csdn.net/qq_37315403/article/details/81485355)。当传递的参数是不可变对象的引用时，虽然传递的是引用，参数变量和原变量都指向同一内存地址，但是不可变对象无法修改，只能复制一份，所以参数的重新赋值不会影响原对象（类似deepcopy），这类似于C语言中的值传递。**既然Python只允许引用传递，那有没有办法可以让两个变量不再指向同一内存地址呢？Python提供了一个copy模块，帮助我们完成这件事。**.......值传递：形参只是得到实参的值，它和实参是两个不同的对象，不会互相影响。引用传递：形参是实参的引用。也就是可以认为形参和实参是同一个对象。     注意是 传递的过程中  赋值 和 （地址的）引用

   值传递（passl-by-value）过程中，被调函数的形式参数作为被调函数的局部变量处理，即在堆栈中开辟了内存空间以存放由主调函数放进来的实参的值，从而成为了实参的一个副本。值传递的特点是被调函数对形式参数的任何操作都是作为局部变量进行，不会影响主调函数的实参变量的值。（被调函数新开辟内存空间存放的是实参的副本值）

   

   引用传递(pass-by-reference)过程中，被调函数的形式参数虽然也作为局部变量在堆栈中开辟了内存空间，但是这时存放的是由主调函数放进来的实参变量的地址。被调函数对形参的任何操作都被处理成间接寻址，即通过堆栈中存放的地址访问主调函数中的实参变量。正因为如此，被调函数对形参做的任何操作都影响了主调函数中的实参变量。（被调函数新开辟内存空间存放的是实参的地址）

   

2. 深浅拷贝都是对源对象的复制，占用不同的内存空间。

   不可变类型的对象，对于深浅拷贝毫无影响，最终的地址值和值都是相等的。

   可变类型： 
   浅拷贝： 值相等，地址相等 
   copy浅拷贝：值相等，地址不相等 (切片赋值是浅拷贝（**切片本身是深拷贝**），只拷贝外侧列表，子列表与原列表公用一份，如果原列表的最外层增删改查，浅拷贝后的列表不变。但是如果原列表的子列表增删改查，浅拷贝后的子列表变化)
   deepcopy深拷贝：值相等，地址不相等



# 二 、装饰器

**问题1.执行顺序**

```python
def decorator_a(**args):
        print("in decorator_a")

        def _(func):
                print("in decorator_a _")

                def wrapper_func(*i, **ki):
                        print("in decorator_a _ wrapper_func")
                        return func(*i,**ki)

                return wrapper_func

        return _


def decorator_b(**args):
        print("in decorator_b")

        def _(func):
                print("in decorator_b _")

                def wrapper_func(*i, **ki):
                        print("in decorator_b _ wrapper_func")
                        return func(*i, **ki)

                return wrapper_func

        return _


@decorator_b(name="myname")
@decorator_a(age=10)
def my(a, b):
        print("in my {0}, {1}".format(a, b))
        return a

my("aaa", {"xxx":"this"})
////////////////////////////////
in decorator_b
in decorator_a
in decorator_a _
in decorator_b _
in decorator_b _ wrapper_func
in decorator_a _ wrapper_func
in my aaa, {'xxx': 'this'}
```

第3行到第7行的输出顺序 可以参考这篇文章 https://segmentfault.com/a/1190000007837364

第1行 第2行顺序原因：

装饰器是为了不改变原来函数代码的基础上，增加新的需求，在python里装饰器就是一个高阶函数。

@是装饰器的语法糖。

例子中的装饰器

```python
@decorator_b(name="myname")
@decorator_a(age=10)
def my(a, b):
        print("in my {0}, {1}".format(a, b))
        return a
        
//////////实际上等同于/////////////
my = decorator_b(name="myname")(decorator_a(age=10)(my))
```

这是立即执行的，所以在例子中最后一行调用my之前，就打印出了1-4行的内容。后面的打印结果没问题，我们就先关注这个赋值语句的执行过程，（对应打印结果的1-4行）

语法有点乱也不慌，用 ast 看看python是如何看待这句话的

```python
import ast
import astunparse

src = '''my = decorator_b(name="myname")(decorator_a(age=10)(my))'''

node = ast.parse(src)
r = astunparse.dump(node)
print(r)
////////////////////结果/////////////////
Module(body=[Assign(
  targets=[Name(
    id='my',
    ctx=Store())],
  value=Call(
    func=Call(
      func=Name(
        id='decorator_b',
        ctx=Load()),
      args=[],
      keywords=[keyword(
        arg='name',
        value=Str(s='myname'))]),
    args=[Call(
      func=Call(
        func=Name(
          id='decorator_a',
          ctx=Load()),
        args=[],
        keywords=[keyword(
          arg='age',
          value=Num(n=10))]),
      args=[Name(
        id='my',
        ctx=Load())],
      keywords=[])],
    keywords=[]))])
```

这时候就很明显了，python认为这是一个Assign赋值语句，赋值给`my`，值是一个Call函数调用的结果，Call的函数本体是这一部分：

```
    func=Call(
      func=Name(
        id='decorator_b',
        ctx=Load()),
      args=[],
      keywords=[keyword(
        arg='name',
        value=Str(s='myname'))]),
```

是`decorator_b(name="myname")`执行的结果！所以这时候打印出了第一行函数的参数是`decorator_a(age=10)(my)`。以此类推，打印出了第二行。

总结：

装饰器就像装箱，装的时候先打包里面的，再打包外层的，打开的时候，先拆外面的再拆里面的，这里b是外层a是里层，这就是3-7行的原因。

但第1行第2行不是装饰器，是生产装饰器的过程，生成顺序按代码中出现的顺序。

```
@decorator_b(name="myname")
@decorator_a(age=10)
```

这里decorator_b(name="myname") 是个函数调用，它的返回值才是作为装饰器（箱子）包装到my函数上，所以它们是在Python解释这个文件时就立刻执行了，解释的顺序自然是从上到下。



# 