## 一、CGO 快速入门

### **1.1、启用 CGO 特性**

在 golang 代码中加入 import “C” 语句就可以启动 CGO 特性。这样在进行 go build 命令时，就会在编译和连接阶段启动 gcc 编译器。

```go
// go.1.15// test1.go
package main
import "C"      // import "C"更像是一个关键字，CGO工具在预处理时会删掉这一行

func main() {
}
```

使用 -x 选项可以查看 go 程序编译过程中执行的所有指令。可以看到 golang 编译器已经为 test1.go 创建了 CGO 编译选项

```go
[root@VM-centos ~/cgo_test/golink2]# go build -x test1.go
WORK=/tmp/go-build330287398
mkdir -p $WORK/b001/
cd /root/cgo_test/golink2
CGO_LDFLAGS='"-g" "-O2"' /usr/lib/golang/pkg/tool/linux_amd64/cgo -objdir $WORK/b001/ -importpath command-line-arguments -- -I $WORK/b001/ -g -O2 ./test1.go    # CGO编译选项
cd $WORK
gcc -fno-caret-diagnostics -c -x c - -o /dev/null || true
gcc -Qunused-arguments -c -x c - -o /dev/null || true
gcc -fdebug-prefix-map=a=b -c -x c - -o /dev/null || true
gcc -gno-record-gcc-switches -c -x c - -o /dev/null || true
.......
```

### **1.2、Hello Cgo**

通过 import “C” 语句启用 CGO 特性后，CGO 会将上一行代码所处注释块的内容视为 C 代码块，被称为**序文（preamble）**。

```text
// test2.go
package main

//#include <stdio.h>        //  序文中可以链接标准C程序库
import "C"

func main() {
    C.puts(C.CString("Hello, Cgo\n"))
}
```

在序文中可以使用 C.func 的方式调用 C 代码块中的函数，包括库文件中的函数。对于 C 代码块的变量，类型也可以使用相同方法进行调用。

test2.go 通过 CGO 提供的 C.CString 函数将 Go 语言字符串转化为 C 语言字符串，最后再通过 C.puts 调用 <stdio.h>中的 puts 函数向标准输出打印字符串。

### **1.3 cgo 工具**

当你在包中引用 import "C"，go build 就会做很多额外的工作来构建你的代码，构建就不仅仅是向 go tool compile 传递一堆 .go 文件了，而是要先进行以下步骤：

1）cgo 工具就会被调用，在 C 转换 Go、Go 转换 C 的之间生成各种文件。

2）系统的 C 编译器会被调用来处理包中所有的 C 文件。

3）所有独立的编译单元会被组合到一个 .o 文件。

4）生成的 .o 文件会在系统的连接器中对它的引用进行一次检查修复。

cgo 是一个 Go 语言自带的特殊工具，可以使用命令 go tool cgo 来运行。它可以生成能够调用 C 语言代码的 Go 语言源文件，也就是说所有启用了 CGO 特性的 Go 代码，都会首先经过 cgo 的"预处理"。

对 test2.go，cgo 工具会在同目录生成以下文件。

```shell
_obj--|
      |--_cgo.o             // C代码编译出的链接库
      |--_cgo_main.c        // C代码部分的main函数
      |--_cgo_flags         // C代码的编译和链接选项
      |--_cgo_export.c      //
      |--_cgo_export.h      // 导出到C语言的Go类型
      |--_cgo_gotypes.go    // 导出到Go语言的C类型
      |--test1.cgo1.go      // 经过“预处理”的Go代码
      |--test1.cgo2.c       // 经过“预处理”的C代码
```

## 二、CGO 的 N 种用法

CGO 作为 Go 语言和 C 语言之间的桥梁，其使用场景可以分为两种：Go 调用 C 程序 和 C 调用 Go 程序。

### 2.1、Go 调用自定义 C 程序

```go
// test3.go
package main

/*
#cgo LDFLAGS: -L/usr/local/lib

#include <stdio.h>
#include <stdlib.h>
#define REPEAT_LIMIT 3              // CGO会保留C代码块中的宏定义
typedef struct{                     // 自定义结构体
    int repeat_time;
    char* str;
}blob;
int SayHello(blob* pblob) {  // 自定义函数
    for ( ;pblob->repeat_time < REPEAT_LIMIT; pblob->repeat_time++){
        puts(pblob->str);
    }
    return 0;
}
*/
import "C"
import (
    "fmt"
    "unsafe"
)

func main() {
    cblob := C.blob{}                               // 在GO程序中创建的C对象，存储在Go的内存空间
    cblob.repeat_time = 0

    cblob.str = C.CString("Hello, World\n")         // C.CString 会在C的内存空间申请一个C语言字符串对象，再将Go字符串拷贝到C字符串

    ret := C.SayHello(&cblob)                       // &cblob 取C语言对象cblob的地址

    fmt.Println("ret", ret)
    fmt.Println("repeat_time", cblob.repeat_time)

    C.free(unsafe.Pointer(cblob.str))               // C.CString 申请的C空间内存不会自动释放，需要显示调用C中的free释放
}
```

CGO 会保留序文中的宏定义，但是并不会保留注释，也不支持#program，**C 代码块中的#program 语句极可能产生未知错误**。

CGO 中**使用 #cgo 关键字可以设置编译阶段和链接阶段的相关参数**，可以使用 ${SRCDIR} 来表示 Go 包当前目录的绝对路径。

使用 C.结构名 或 C.struct_结构名 可以在 Go 代码段中定义 C 对象，并通过成员名访问结构体成员。

test3.go 中使用 C.CString 将 Go 字符串对象转化为 C 字符串对象，并将其传入 C 程序空间进行使用，由于 C 的内存空间不受 Go 的 GC 管理，因此需要显示的调用 C 语言的 free 来进行回收。详情见第三章。