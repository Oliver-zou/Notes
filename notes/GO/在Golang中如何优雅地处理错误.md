## 一 、Golang中并不那么优雅的错误处理

### 1.1“对开发人员具有一定的心智负担”

在一些文章[1]中，把Go的错误处理形容为“嘈杂”的，甚至“令人作呕”的——你需要为所有可能出现的错误做显式的处理：

```go
func first() error {return nil}
func second() error {return nil}
func third() error {return nil}
func fourth() error {return nil}
func fifth() error {return nil}

func Do() error {
    var err error
    if err = first(); err == nil {
        if err = second(); err == nil {
            if err = third(); err == nil {
                if err = fourth(); err == nil {
                    if err = fifth(); err == nil {
                        return nil
                    }
                }
            }
        }
    }
    return err
}
```

当你不得不为每一条字段的校验、每一个方法的调用、每一次数据库的查询……都要去判断其返回的err是否为nil时，你会发现这个过程就像是Russ Cox(Go语言作者之一)所说的：“**这对开发人员来说可能有一定的心智负担**”。

### 1.2流派众多

为了降低开发人员的“心智负担”，在Go社区出现了多种错误处理的流派。

Rob Pike(Go语言作者之一)拿出了看似补救，但实则被认为更为危险的“创可贴”；无独有偶，Andrew Gerrand(同为Go语言作者之一)在[《Error handling and Go》](https://go.dev/blog/error-handling-and-go)一文中也给出了他的处理过程；遑论还有来自社区更接地气的`panic-defer-recover`方法等。

打着”少即是多“这一旗号的Go在错误处理方面却显得纷繁芜杂，莫衷一是。

### 官方的考量

为什么Go不像其他大多数的高级语言那样，提供`try-catch`语法呢？要知道哪怕是Rust，也提供了Panic!宏。

可能同样来自Russ Cox的回答比较能够代表官方的考量：“**这种基于[comma-ok断言]的错误处理机制比[try-catch]更适用于大型软件**“。按照Go作者们的理念，他们有意识地想要把go语言设计成一个使用显式错误返回和进行显式错误处理的语言。

纵观错误处理的历史[2]：C时代的变量传递或全局变量的方式属实不便；C++时代的try-catch则让调用方心力交瘁，因为你无法确认一个方法是否会抛出异常；Java对异常进行了显示声明，使得错误处理大有改观，但也难免逐渐被滥用成更像是一种流程控制。对于Go，它解决问题的方式就是“取消异常”， 而通过多值返回error来表示错误。对于真正panic的异常：“When you panic in Go, you’re freaking out, it’s not someone elses problem, it’s game over man”。

不过幸好，Go的作者们终于像接受了泛型那样，接受了`try-catch`的思想。在Go2.0草案中[关于错误处理](https://go.googlesource.com/proposal/+/master/design/go2draft-error-handling-overview.md)的章节中，已经打算引入handle和check关键字：

```go
func main() {
	handle err {
		log.Fatal(err)
	}
	hex := check ioutil.ReadAll(os.Stdin)
	data := check parseHexdump(string(hex))
	os.Stdout.Write(data)
}
```

这释放了一个非常积极的信号，不过其真正落地可能也需要数年之久。在期待的同时，现在的我们该如何优雅地进行错误处理呢？接下来先看看业界都有哪些错误处理的流派：

# 二、常见的错误处理流派

各流派之间的差异主要体现在错误的产生、流程的控制、错误的返回三个方面：

## 2.1. if err != nil return err 流

最正宗的流派自然是官方推荐的`if err != nil return err`流：

- 通过`fmt.Errorf`、`errors.New`等方法产生原生或自定义错误
- 在发现`err != nil`之后直接返回，不执行剩下的函数过程
- 错误通过返回值直接返回

> 在go1.13之后，官方引入了对错误类型断言更友好的机制：对`fmt.Errorf`添加`warping`功能；在errors包中提供了`Is`和`As`函数。在实际应用中，通过该机制可以非常方便地进行error嵌套和判断，相关内容可以非常方便地在各大搜索引擎中获取，本文在此不加赘述。

## 2.2. panic-defer-recover 流

该流派也有着一定的拥簇，其主要特点为：

- 使用`panic`来把错误抛出，同样不执行剩下的函数过程
- 在函数头通过defer来把`panic`的异常捕捉，然后转为返回值返回

见如下伪代码：

```go
func Demo() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}
	// ...
	sErr := makeError()
	if sErr != nil {
		panic(sErr)
	}
}
```

事实上，使用该流派者主要是为了对冗余的错误处理过程进行简写，其中较为成熟的项目有goconvey等。以上面的伪代码为例，可以进一步优化：

```go
func Demo() (err error) {
	defer HandleErr(&err)
	// ...
	ThrowErr(makeReturnError())
}

func HandleErr(err *error) {
	if r := recover(); r != nil {
		*err = r.(error)
	}
}
func ThrowErr(err error) {
	if err != nil {
		panic(err)
	}
}
```

把错误的抛出和处理分别放到公共函数ThrowErr和HandleErr中，就可以大大简写Demo的错误处理过程。当然，在实际实现中还要考虑如何放过原生的panic等的逻辑，这里不作展开。

## 2.3. errors-are-values 流

这个流派来自Go语言之父Rob Pike的博客中的同名文章，其主要特点为：

- 错误的产生与否不影响主流程的执行过程
- 在主流程的结尾处进行错误判断并返回

见如下伪代码：

```go
type Worker struct {
	...
    err error
}
func (w *Worker) DoJob(...) {
    if w.err != nil {
        return // 当已经出错时，就不做任何处理
    }
    // 正常业务逻辑
	...
}

func try() error {
    w := &Worker{...}
    w.DoJob(a)
    w.DoJob(b)
    ...
    if w.err != nil {
        return w.err
    }
}
```

把error当作receiver的一个成员变量，并且在receiver的每一个method中，判断成员error是否为空，为空才处理接下来的逻辑。

#  三、规范错误处理方案

## 3.1Golang错误处理规范

- error必须通过返回值返回，必须进行显式处理(或赋值给明确忽略)
- 有多个返回值时，error必须为最后一个返回值

```go
func do() (error, int) {} // Not Good
func do() (int, error) {} // Good
```

- error内容不需要标点结尾
- 采用独立的错误流进行处理

```go
// Not Good
if err != nil {
	// Handle error
} else {
	// Normal Code
} 
// Good
if err != nil {
	// Handler error
	return
}
// Normal Code
```

- 错误的判断应独立处理，不参与其他逻辑判断

```go
// Not Good
if err != nil || code != 200 {
	return err
}
// Good
if err != nil {
	return err
}
if code != 200 {
	return fmt.Errorf("code error")
}
```

- 推荐使用go1.13+的`warping`特性
- 在业务逻辑处理中禁止使用`panic`
- 对于包，可导出的接口一定不能有`panic`；在包内传递错误时，不推荐使用`panic`
- 建议在`main`包中使用`log.Fatal`来记录错误
- `panic`的捕捉只能到`goroutine`的最顶层，亦即在每个自行启动的`goroutine`的入口处捕获`panic`来进行处理

## 3.2. 部分场景中的考量

但在个别场景中，使用`panic`来进行错误传递会有着非常大的收益。比如单元测试、低频但繁冗的导入数据校验、对性能要求不高但重视代码可读性等。

设有如下代码：

```go
// ValidateStaffInfo 校验导入的员工信息是否正确
func ValidateStaffInfo(s StaffInfo) error {
	if "" == s.EnglishName {
		return fmt.Errorf("员工英文名不能为空")
	}
	if 0 == s.StaffId {
		return fmt.Errorf("员工帐号ID不能为空")
	}
	staffTypes := []string{ "正式员工", "外包员工", "物业人员", "其他"}
	chkStaffType := false
	for _, sType := range staffTypes {
		if sType == s.StaffType {
			chkStaffType = true
			break
		}
	}
	if !chkStaffType {
		return fmt.Errorf("员工类型%s不正确", s.StaffType)
	}
	err := ValidateStaffDept(s)
	if nil != err {
		return err
	}
	return nil
}
```

上面的函数来自某系统中，用户通过Excel导入数据时的校验过程。该过程并不参与主要的业务逻辑，使用的频率不大，性能要求不高，可以做如下优化以减轻开发的负担和提高代码可读性：

```go
// ValidateStaffInfo 校验导入的员工信息是否正确
func ValidateStaffInfo(s StaffInfo) (err error) {
	defer xerr.HandleErr(&err)
	xerr.ThrowErrMsgWhen("" == s.EnglishName, "员工英文名不能为空")
	xerr.ThrowErrMsgWhen(0 == s.StaffId, "员工帐号ID不能为空")
	xerr.ThrowErrMsgWhenNotExists(
		s.StaffType, []string{ "正式员工", "外包员工", "物业人员", "其他"},
		"员工类型%s不正确", s.StaffType)
	xerr.ThrowErr(ValidateStaffDept(s))
	return nil
}
```

其中的`xerr`是对前文小节中`panic-defer-recover`思想的实现和包装。

不过在实现中也需要考虑一些细节上的处理：

- 要注意空指针取值的问题。比如`xerr.ThrowErrMsgWhen(ptr == nil, "引用的对象不存在:%s", ptr.Name)`会在ptr为nil的时候必然报错；
- 要注意对原生`panic`的处理。因为使用了`recover`，所以捕捉到的`panic`也包含了原生抛出的，比如除0，空指针取值等。这里建议`xerr.ThrowErrXXX`抛出的错误为自定义错误，然后在`HandleErr`时通过`errors.Is`或`errors.As`来进行判断，如果不是则重新`panic`；
- 对标准错误处理接口的适配。比如提供`Error`、`Format`、`Warp`、`Unwarp`、`Is`、`As`等接口和支持堆栈输出等，这一部分可以参考官方的`errors`包，这里就不展开了。

## 3.3. 进一步探讨

如果有了解过Go的语言设计和实现，那么可以看到：Go语言跟C不同，不是使用栈指针寄存器和栈基址寄存器确定函数的栈的，而是基于``runtime._panic``这个数据结构。通过该结构，每个`goroutine`都有着自己的栈，并且还能通过连续栈(continuous stack)技术来进行自动扩缩容。换而言之，使用`panic-defer-recover`相较于通过返回值返回会有一定的额外开销，但是也不至于像其他语言(C++,JAVA,C#等)中的`try-catch`那么大。

为了量化使用`panic-defer-recover`带来的性能影响，这里笔者进行了个小实验。

先预设两个概念：

- 嵌套调用：函数A被函数B调用，那么B就是对A的一次嵌套；函数C再对B进行调用，那么C就是对A的二次嵌套；
- `panic`重放：在`recover`捕捉到了错误之后，再重新通过`panic`向上抛出。

在配置为i7-4790 CPU @ 3.60GHz / win10x64 / 16G内存的机器上，得到如下结果：

**执行耗时**：

| 错误处理的方式  | 无嵌套，循环千万次耗时 | 嵌套10层，循环千万次耗时 | 嵌套100层，循环千万次耗时 |
| :-------------: | :--------------------: | :----------------------: | :-----------------------: |
|   返回值返回    |         440ms          |          640ms           |          5040ms           |
|  panic(无重放)  |         3790ms         |          8370ms          |          42660ms          |
| panic(逐层重放) |        13320ms         |     372360ms(6.2min)     |     23952380ms(6.7h)      |

**基准测试**：

| 错误处理的方式 |             无嵌套             |            嵌套10层             |             嵌套100层             |
| :------------: | :----------------------------: | :-----------------------------: | :-------------------------------: |
|   返回值返回   | 34.5 ns/op 16 B/op 1 allocs/op | 60.0 ns/op 16 B/op 1 allocs/op  |   453 ns/op 16 B/op 1 allocs/op   |
| panic(无重放)  | 600 ns/op 16 B/op 1 allocs/op  | 2976 ns/op 16 B/op 1 allocs/op  |  27113 ns/op 16 B/op 1 allocs/op  |
| panic(有重放)  | 1305 ns/op 16 B/op 1 allocs/op | 37559 ns/op 16 B/op 1 allocs/op | 2729109 ns/op 30 B/op 1 allocs/op |

- 结果说明：34.5 ns/op 16 B/op 1 allocs/op 表平均一次执行耗时34.5ns，占用内存16字节，进行了1次内存分配

## 3.4. 小结

通过实验可知，当使用`panic-defer-recover`时：

- 在逐层重放的情况下，会对性能有较大的影响；
- 得益于其精巧的连续栈设计，在内存消耗方面没有明显的提高。

结合对规范践行的考量有如下结论：

- 在大多数场景下，尽量遵循通过返回值显式返回错误的官定规则；
- 在不计较性能，并对代码可读性和减轻开发负担有所要求的场景下，可以使用`panic-defer-recover`来处理业务错误；
- 使用`panic-defer-recover`时要特别注意它的局限性：
  - 相比非`panic`会多出数十倍时间的消耗；
  - 容易产生空指针取值的问题；
  - 要注意对原生`panic`的处理；
  - 要注意对标准错误接口的适配。

# 四、扩展：进一步的思考

以上的讨论都局限在代码和逻辑层面，而对于一门后台开发语言，进行错误处理往往要考虑到如下三个维度[3]：

- **函数内部的错误处理**：这是一个函数在执行中遇到各种错误的处理过程。
- **服务内部的错误处理**：这是一个服务在进行一个或多个函数的调用时，对错误的统一管理过程。
- **面向外部的错误处理**：这是服务在某个业务处理失败时，如何返回友好的错误信息，让调用方更好地理解和处理的过程。

除了函数内部的错误处理，还要考虑在整个服务内部的统一错误管理：如何对错误进行编码，嵌套的错误该如何处理，错误的日志与跟踪等；而当面向外部时，更要考虑一个错误是如何被消费的：前端给用户的提示是否更贴合业务，堆栈信息的显示与否，错误提示信息的国际化等等。



#### 引用

[1] Donng.Go 语言的优点，缺点和令人厌恶的设计[EB/OL].https://studygolang.com/articles/12907

[2] Dave Cheney.Why Go gets exceptions right[EB/OL].https://dave.cheney.net/2012/01/18/why-go-gets-exceptions-right

[3] andruzhang.一套在 Go 中优雅地传递、返回、暴露错误，同时便于回溯翻查的解决方案[EB/OL].[https://km.woa.com/group/34868/articles/show/486544](https://km.woa.com/group/34868/articles/show/486544?kmref=home_recommend_read)