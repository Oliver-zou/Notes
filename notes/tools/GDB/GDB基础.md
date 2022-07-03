### 一、调试准备

1.1 编译时加入调试信息：-g 选项
1.2 关闭编译器的程序优化选项。程序优化选项，一般有五个级别，从 O0 ~ O4，O0 表示不优化，从 O1 ~ O4 优化级别越来越高。
1.3 调试完之后可以移除调试信息：

- 编译时不加 -g 选项
- 使用 linux 的 strip 命令移除掉某个程序中的调试信息

### 二、使用方式

#### 2.1 直接调试目标程序：`gdb filename`

#### 2.2 附加进程：`gdb -p pid`

当用 gdb attach 上目标进程后，调试器会暂停进程，此时可以使用 continue 命令让程序继续运行。当调试完程序想结束此次调试时，而且想让当前进程继续运行，则可以用 detach 命令让 GDB 调试器与程序分离。

#### 2.3  调试 core 文件：`gdb filename corename`

（1）让程序在崩溃的时候产生 core 文件，这样就可以使用这个 core 文件来定位崩溃的原因。Linux 系统默认不开启生成 core 文件，可以使用`ulimit -a`来查看 core file size 的值。默认是0表示不开启。则可以使用`ulimit -c unlimited`来开启。不过这样修改，当我们关闭这个会话时，设置项的值就又会恢复成0。所以让这个选项永久生效的方式是把`ulimit -c unlimited`这一行加到 `/etc/profile`文件中去。
（2）/proc/sys/kernel/core_pattern 可以设置格式化的 core 文件保存位置或文件名。比如：

```shell
echo "/corefile/core-%e-%p-%t" > /proc/sys/kernel/core_pattern
```

| 参数名称 |           含义            |
| :------: | :-----------------------: |
|    %p    |            pid            |
|    %u    |            uid            |
|    %t    | core 文件生成时间（UNIX） |
|    %h    |          主机名           |
|    %e    |          程序名           |

### 三、常用调试命令

|  命令名称   | 命令缩写 |                        命令说明                        |
| :---------: | :------: | :----------------------------------------------------: |
|     run     |    r     |                          运行                          |
|  continue   |    c     |                  让暂停的程序继续运行                  |
|    next     |    n     |                      运行到下一行                      |
|    step     |    s     |  如果有调用函数，进入调用的函数内部，相当于 step into  |
|    until    |    u     |                   运行到指定行停下来                   |
|   finish    |    fi    |          结束当前调用函数，到上一层函数调用处          |
|   return    |  return  |    结束当前调用函数并返回指定值，到上一层函数调用处    |
|    jump     |    j     |           将当前程序执行流跳转到指定行或地址           |
|    print    |    p     |                   打印变量或寄存器值                   |
|  backtrace  |    bt    |                 查看当前线程的调用堆栈                 |
|    frame    |    f     | 切换到当前调用线程的指定堆栈，具体堆栈通过堆栈序号指定 |
|   thread    |  thread  |                     切换到指定线程                     |
|    break    |    b     |                        添加断点                        |
|   tbreak    |    tb    |                      添加临时断点                      |
|   delete    |   del    |                        删除断点                        |
|   enable    |  enable  |                      启用某个断点                      |
|   disable   |  disabl  |                      禁用某个断点                      |
|    watch    |  watch   |        监视某一个变量或内存地址的值是否发生变化        |
|    list     |    l     |                        显示源码                        |
|    info     |    i     |                  查看断点/线程等信息                   |
|    ptype    |  ptype   |                      查看变量类型                      |
| disassemble |   dis    |                      查看汇编代码                      |
|  set args   |          |                 设置程序启动命令行参数                 |
|  show args  |          |                  查看设置的命令行参数                  |

#### 查看源码 list/directory

1.list 可以显示当前断点处的代码，也可以显示其他文件某一行的代码。

2.list 命令会显示当前前后的10行代码，继续输入会继续往后显示10行。

3.也可以指定是往前还是往后显示代码，命令分别是 `list +` 和 `list -`。

4.用 directory 指定源码目录：`directory [path]`

5.`show directories` 显示源文件搜索路径。

#### 打印 print/ptype

1.print 命令不仅可以显示变量值，还可以显示进行一定运算的表达式计算结果值，还可以显示一些函数的执行结果值。

2.某个时刻，某个系统函数执行失败了，通过系统变量 errno 得到一个错误码，则可以使用`p strerror(errno)`将这个错误码对应的文字信息打印出来，这样就不用费劲地去查 man 手册了。

3.当使用 print 命令打印一个字符串或者字符数组时，如果该字符串太长，会默认显示不全。可以通过`set print element 0`进行设置，使得可以完整地显示字符串。

4.让打印更美观 `set print pretty on`

5.打印某个对象指针 `print *ptr`

6.ptype 就是"print type"，作用是输出变量的类型。

7.对于一个复合数据类型的变量，ptype 不仅列出了这个变量的类型，而且详细地列出了每个成员变量的字段名。

8.`ptype[/FLAGS] TYPE-NAME | EXPRESSION`
参数可以是由 typedef 定义的类型名， 或者 struct STRUCT-TAG 或者 class CLASS-NAME 或者 union UNION-TAG 或者 enum ENUM-TAG。根据所选的栈帧的词法上下文来查找该名字。

9.类似的命令是 whatis，区别在于 whatis 不展开由 typedef 定义的数据类型，而 ptype 会展开，举例如下：

```c++
/* 类型声明与变量定义 */
typedef double real_t;
struct complex {
    real_t real;
    double imag;
};
typedef struct complex complex_t;
complex_t var;
real_t *real_pointer_var;
```

这两个命令给出了如下输出：

```c++
(gdb) whatis var
type = complex_t
(gdb) ptype var
type = struct complex {
    real_t real;
    double imag;
}
(gdb) whatis complex_t
type = struct complex
(gdb) whatis struct complex
type = struct complex
(gdb) ptype struct complex
type = struct complex {
    real_t real;
    double imag;
}
(gdb) whatis real_pointer_var
type = real_t *
(gdb) ptype real_pointer_var
type = double *
```

10.`set print address on` 打开地址输出，当程序显示函数信息时，GDB会显出函数的参数地址。默认是打开的。

11.`set print array on` 打开数组显示，打开后当数组显示时，每个元素占一行，如果不打开的话，每个元素则以逗号分隔。

12.`set print elements` 这个选项主要是设置数组的，如果你的数组太大了，那么就可以指定一个来指定数据显示的最大长度，当到达这个长度时，GDB就不再往下显示了。如果设置为0，则表示不限制。默认是200个。

13.`set print null-stop` 如果打开了这个选项，那么当显示字符串时，遇到结束符则停止显示。这个选项默认为off。

14.`set print union` 显示结构体时，是否显式其内的联合体数据。

15.`set print object` 在C++中，如果一个对象指针指向其派生类，如果打开这个选项，GDB会自动按照虚方法调用的规则显示输出，如果关闭这个选项的话，GDB就不管虚函数表了。

16.记不清楚的话，就用help命令：`help set print`

17.`p <addr>@<len>`
有时候，你需要查看一段连续的内存空间的值。比如数组的一段，或是动态分配的数据的大小。你可以使用GDB的“@”操作符，“@”的左边是第一个内存的地址的值，“@”的右边则你你想查看内存的长度。例如，你的程序中有这样的语句：
`int *array = (int *) malloc (len * sizeof (int));`
于是，在GDB调试过程中，你可以以如下命令显示出这个动态数组的取值：
`p *array@len`
@的左边是数组的首地址的值，也就是变量array所指向的内容，右边则是数据的长度，其保存在变量len中，其输出结果，大约是下面这个样子的：

```python
(gdb) p *array@len
$1 = {2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40}
```

如果是静态数组的话，可以直接用print数组名，就可以显示数组中所有数据的内容了。

18.一般来说，GDB会根据变量的类型输出变量的值。但你也可以自定义GDB的输出的格式。例如，你想输出一个整数的十六进制，或是二进制来查看这个整型变量的中的位的情况。要做到这样，你可以使用GDB的数据显示格式：

| 格式 |              含义              |
| :--: | :----------------------------: |
|  x   |    按十六进制格式显示变量。    |
|  d   |     按十进制格式显示变量。     |
|  u   | 按十六进制格式显示无符号整型。 |
|  o   |     按八进制格式显示变量。     |
|  t   |     按二进制格式显示变量。     |
|  a   |    按十六进制格式显示变量。    |
|  c   |      按字符格式显示变量。      |
|  f   |     按浮点数格式显示变量。     |

比如：

```c++
(gdb) p i
$21 = 101
(gdb) p/a i
$22 = 0x65
(gdb) p/c i
$23 = 101 'e'
(gdb) p/f i
$24 = 1.41531145e-43
(gdb) p/x i
$25 = 0x65
(gdb) p/t i
$26 = 1100101
```

#### 信息查看 info

1.`info break` 可以查看当前断点。

2.`info thread` 可以查看当前进程有哪些线程。

线程的 id 是第三栏括号里的内容（如 LWP 21723）中的21723。

3.`info args` 可以查看当前函数的参数值。

4.`info lolcals` 打印出当前函数中所有局部变量及其值。

5.`info registers` 查看所有寄存器的情况。（包括浮点寄存器）
同样可以使用print命令来访问寄存器的情况，只需要在寄存器名字前加一个$符号就可以了。如：`p $eip`

6.`info f` 打印出当前栈层的信息

7.`i line` 查看源代码在内存中的地址。后面可以跟“行号”、“函数名”、“文件名:行号”、“文件名:函数名”

#### 单步调试 next/step

1.`next` 表示 step over 单步步过，即遇到函数调用直接跳过，不进入函数体内部。

2.`ni` 是汇编指令级别的下一条

3.`step` 表示 step into 单步步入，遇到函数调用，进入函数内部。

4.`si` 表示汇编指令级别的进入函数内部。

5.`step [N]` 参数 N 表示执行 N 次（或由于另一个原因直到程序停止）。

#### 程序执行 finish/return/until/jump

1.实际调试时，在某个函数中调试一段时间后，不需要再一步步执行到函数返回处，使用 `finish` 命令可以直接执行完当前函数并回到上一层调用处。

2.`return` 命令是直接结束执行当前函数，还可以指定该函数的返回值。也就是说，如果当前函数还有剩余的代码未执行完毕，也不会执行了。

3.`until` 命令可以指定程序运行到某一行停下来。
`until <location>` 或 `u <location>` 继续运行程序，直到达到指定的位置，或者当前栈帧返回，循环结束。

4.`continue`在信号或断点之后，继续运行被调试的程序。
`continue [N]` 如果从断点开始，可以使用数字 N 作为参数，这意味着将该断点的忽略计数设置为 N - 1(以便断点在第 N 次到达之前不会中断)。如果启用了非停止模式（使用 show non-stop 查看），则仅继续当前线程，否则程序中的所有线程都将继续。

5.`jump <location>` 中 location 可以是程序的行号或者函数的地址，jump 会让程序执行流跳转到指定位置执行，当然其行为是不可控制的，因为会直接忽略中间的代码。

- 如果 jump 跳转到的位置后续没有断点，那么 GDB 会执行完跳转处的代码后会继续执行下去。
- jump 命令的一个妙用就是可以执行一些我们想要执行的代码，而这些代码在正常的逻辑下可能并不会执行。

#### 查汇编指令 disassemble

1.disassemble 命令可以查看某段代码的汇编指令。

- `disassemble func`
- `disassemble /rm` 查看当前函数的汇编信息
- /m指示显示汇编指令的同时，显示相应的程序源码
- /r指示显示十六进制的 raw instructions

2.GDB 默认反汇编为 AT&T 格式的指令，可以通过`show disassembly-flavor`查看。如果习惯 intel 汇编格式可以用命令`set disassembly-flavor intel`来设置。

3.可以使用 `info line` 命令来查看源代码在内存中的地址。Info line 后面可以跟“行号”,“函数名”,“文件名:行号”,“文件名:函数名”,这个命令会打印出所指定的源代码在运行时的内存地址,如:

```shell
(gdb) info line tst.c:func
Line 5 of "tst.c" starts at address 0x8048456 
```

#### 设置运行参数 set args 和 show args

1.在用 GDB attach 程序后，在使用 run 命令之前，使用`set args 参数内容`来设置命令行参数。

2.如果单个命令行参数之间含有空格，可以使用引号将参数包裹起来。

3.如果想清除掉已经设置好的命令行参数，使用 set args 不加任何参数即可。

#### 普通断点 break 和 tbreak

GDB 中支持3类断点，分别为 `break` 普通断点、`watch` 观察断点、`catch` 捕捉断点。

1.使用 `break` 命令（缩写 b）来设置断点。

2.`tbreak` 命令也是添加一个断点，只是这个断点是临时的。所谓临时断点，就是一旦该断点触发一次后就会自动删除。

3.`break` 当不带参数时，在所选栈帧中执行的下一条指令处设置断点。

4.`break FUNCTION`
在某个函数上设置断点。函数重载时，有可能同时在几个重载的函数上设置了断点 ，在 C++ 中可以使用 class::function 或 function(type, …) 格式来指定函数名。

5.`break +OFFSET`
`break -OFFSET`
在当前程序运行到的前几行或后几行设置断点。

6.`break LINENUM`在行号为LINENUM的行上设置断点

7.`break <filename:linenum>` 在源码文件 filename 的 linenum 行处打断点。

8.`break <filename:function>` 在源码文件 filename 的 function 函数入口处打断点。

9.`break <address>` 在程序指令的地址处打断点。

10.`break ... if <cond>` 设置条件断点，… 代表上述参数之一（或无参数），cond为条件表达式，仅在 cond 值非零时暂停程序执行。

11.`enable 断点编号`
恢复暂时失活的断点，要恢复多个编号的断点，可用空格将编号分开

12.`disable 断点编号`
使断点失效，但是断点还在

13.`delete 断点编号或者表达式`
删除某断点

#### 监视断点 watch

1.watch 系有三类：

- `watch` ：可以用来监视一个变量或者一段内存。当这个变量或者该内存处的值发生变化时，GDB 就会中断下来。被监视的某个变量或者某个内存地址会产生一个 watch point （观察点）
- `rwatch` ：只要程序中出现读取目标变量（表达式）的值的操作，就会停止运行
- `awatch` ：只要程序中出现读或者改目标变量（表达式）的值的操作，就会停止运行

2.watch 的使用方式是`watch 变量名或内存地址`

3.设置的观察点有两类：

- 软件观察点：即用 watch 命令监控目标变量（表达式）后，GDB 会以单步执行的方式运行程序，每步执行后都会检测观察点的值是否发布变化，如果改变了则程序停止运行
- 硬件观察点：是用少量寄存器来作为观察点，因为寄存器个数有限资源宝贵，所以硬件观察点数量也有限，并且也没法给占用字节数较多的目标变量（表达式）设置硬件观察点，可以通过 `show can-use-hw-watchpoints` 来查看当前环境是否支持硬件观察点

#### 捕捉断点 catch

1.捕捉断点的作用是，监控程序中某一事件的发生，例如程序发生某种异常时、某一动态库被加载时等等，一旦目标事件发生，则程序停止执行。

2.通过 `catch event` 来建立捕捉断点，event 参数表示要监控的具体事件，部分事件如下：

|      event事件       |                             含义                             |
| :------------------: | :----------------------------------------------------------: |
|  throw [exception]   | 当程序中抛出 exception 指定类型异常时，程序停止执行。如果不指定异常类型（即省略 exception），则表示只要程序发生异常，程序就停止执行 |
|  catch [exception]   | 当程序中捕获到 exception 异常时，程序停止执行。exception 参数也可以省略，表示无论程序中捕获到哪种异常，程序都暂停执行 |
| load/unload [regexp] | regexp 表示目标动态库的名称，load 命令表示当 regexp 动态库加载时程序停止执行；unload 命令表示当 regexp 动态库被卸载时，程序暂停执行。regexp 参数也可以省略，此时只要程序中某一动态库被加载或卸载，程序就会暂停执行 |

更多事件可以通过 `help catch` 来查看。

3.跟 `break` 有个对应的 `tbreak` 表示临时断点一样，`catch` 也有对应的 `tcatch` 表示只监控一次，触发后就失效。

#### 修改条件断点 condition

1.`condition` 命令可以为现有的断点添加条件表达式使之成为条件断点，也可以对条件断点的条件表达式进行修改或者删除。

2.`condition bnum expression` ： bnum 表示目标断点的编号（通过 `i b` 可以查看目前的断点列表），expression 表示条件表达式。这个命令表示为断点 bnum 添加或修改成条件表达式 expression。比如 `condition 1 num==3` 表示当 `num==3` 成立时触发断点1。

3.`condition bnum` ： 表示删除断点 bnum 的条件表达式，使之成为普通的无条件断点。

#### 自动打印变量 display

1.`display` 命令用来监视变量或者内存地址，每次程序中断下来都会自动输出这些变量或内存的值。

2.使用`info display`查看当前已经自动添加了哪些值。

3.使用`delete display`清除全部需要自动输出的变量。

4.使用`delete display 编号`删除某个自动输出的变量。

#### 堆栈相关：bt、f

1.`bt`
显示所有的调用栈帧。该命令可用来显示函数的调用顺序。

2.`bt <n>`
n 是一个正整数,表示只打印栈顶上 n 层的栈信息。

3.`bt <-n>`
-n 是一个负整数,表示只打印栈底下 n 层的栈信息。

4.如果你要查看某一层的信息,你需要切换当前的栈,一般来说,程序停止时,最顶层的栈就是当前的栈,如果你要查看栈下面层的详细信息,首先要做的是切换当前栈。`frame <n>`，`f <n>`n 是一个从 0 开始的整数,是栈中的层编号。比如:frame 0,表示栈顶,frame ,表示栈的第二层。
5.`up <n>`
表示向栈的上面移动 n 层,可以不打 n,表示向上移动一层。

6.`down <n>`
表示向栈的下面移动 n 层,可以不打 n,表示向下移动一层。

7.`info f`
这个命令会打印出更为详细的当前栈层的信息,只不过,大多数都是运行时的内存地址 。
比如函数的地址,调用函数的地址,被调用函数的地址,目前的函数是由什么样的程序语言写成的、函数参数地址及值,局部变量的地址等等。如:

```shell
(gdb) info f
Stack level 0, frame at 0xbffff5d4:
eip = 0x804845d in func (tst.c:6); saved eip 0x8048524
called by frame at 0xbffff60c
source language c.
Arglist at 0xbffff5d4, args: n=250
Locals at 0xbffff5d4, Previous frame's sp is 0x0
Saved registers:
ebp at 0xbffff5d4, eip at 0xbffff5d8
```

#### 查看内存 x

`x/<n/f/u> <addr>`
n、f、u 是可选参数，用于指定要显示的内存以及如何格式化。addr 是要开始显示内存的地址的表达式。
n 是一个正整数，表示需要显示的内存单元的个数，也就是说从当前地址向后显示几个内存单元的内容，一个内存单元的大小由后面的u定义。
f 显示格式（初始默认值是 x），显示格式是 print(‘x’，‘d’，‘u’，‘o’，‘t’，‘a’，‘c’，‘f’，‘s’) 使用的格式之一，再加 i（机器指令）。
u 单位大小，表示从当前地址往后请求的字节数，如果不指定的话，GDB默认是4个bytes。u参数可以用下面的字符来代替，b 表示单字节，h 表示双字节，w 表示四字节，g 表示八字节。
当我们指定了字节长度后,GDB 会从指定的内存地址开始,读写指定字节,并把其当作一个值取出来。
比如:
`x/3uh 0x54320` 表示从地址 0x54320 开始以无符号十进制整数的格式，双字节为单位来显示 3 个内存值。

`x/16xb 0x7f95b7d18870` 表示从地址 0x7f95b7d18870 开始以十六进制整数的格式，单字节为单位显示 16 个内存值。

#### 窗口布局 layout

layout 用于分割窗口，可以一边查看代码，一边测试。主要有以下几种用法：

`layout src`：显示源代码窗口

`layout asm`：显示汇编窗口

`layout regs`：显示源代码/汇编和寄存器窗口

`layout split`：显示源代码和汇编窗口

`layout split`：显示源代码和汇编窗口

`layout next`：显示下一个layout

`layout prev`：显示上一个layout

Ctrl + L：刷新窗口

Ctrl + x，再按1：单窗口模式，显示一个窗口

Ctrl + x，再按2：双窗口模式，显示两个窗口

Ctrl + x，再按a：回到传统模式，即退出layout，回到执行layout之前的调试窗口。

#### 环境变量

1.你可以在GDB的调试环境中定义自己的变量，用来保存一些调试程序中的运行数据。GDB的环境变量和UNIX一样，也是以$起头，如：`set $foo = *object_ptr`
使用环境变量时，GDB会在你第一次使用时创建这个变量，而在以后的使用中，则直接对其賦值。环境变量没有类型，你可以给环境变量定义任一的类型。包括结构体和数组。

2.`show convenience` 查看当前所设置的所有的环境变量。

### 四、调试技巧

#### 带参数运行目标二进制

```
gdb --args <target_bin> <args>
```

#### 让被 GDB 调试的程序接收信号

1.`info signals signal` 可以打印当前 GDB 调试器对指定 signal 的信号处理方式，如果省略 signal 则会打印全部信号，如下：

```shell
(gdb) i signals 
Signal        Stop      Print   Pass to program Description

SIGHUP        Yes       Yes     Yes             Hangup
SIGINT        Yes       Yes     No              Interrupt
SIGQUIT       Yes       Yes     Yes             Quit
SIGILL        Yes       Yes     Yes             Illegal instruction
SIGTRAP       Yes       Yes     No              Trace/breakpoint trap
SIGABRT       Yes       Yes     Yes             Aborted
SIGEMT        Yes       Yes     Yes             Emulation trap
SIGFPE        Yes       Yes     Yes             Arithmetic exception
SIGKILL       Yes       Yes     Yes             Killed
SIGBUS        Yes       Yes     Yes             Bus error
SIGSEGV       Yes       Yes     Yes             Segmentation fault
SIGSYS        Yes       Yes     Yes             Bad system call
SIGPIPE       Yes       Yes     Yes             Broken pipe
SIGALRM       No        No      Yes             Alarm clock
SIGTERM       Yes       Yes     Yes             Terminated
SIGURG        No        No      Yes             Urgent I/O condition
SIGSTOP       Yes       Yes     Yes             Stopped (signal)
SIGTSTP       Yes       Yes     Yes             Stopped (user)
SIGCONT       Yes       Yes     Yes             Continued
SIGCHLD       No        No      Yes             Child status changed
```

2.`handle signal mode` 可以改变 GDB 信号处理的设置：
（1）signal 参数表示要设定的目标信号，它通常为某个信号的全名（SIGINT）或者简称（如 INT）；如果要指定所有信号，可以用 all 表示
（2）mode 参数用于明确 GDB 处理该目标信息的方式，其值可以是如下几个的组合：

- nostop：当信号发生时，GDB 不会暂停程序，其可以继续执行，但会打印出一条提示信息，告诉我们信号已经发生
- stop：当信号发生时，GDB 会暂停程序执行
- noprint：当信号发生时，GDB 不会打印出任何提示信息
- print：当信号发生时，GDB 会打印出必要的提示信息
- nopass（或者 ignore）：GDB 捕获目标信号的同时，不允许程序自行处理该信号
- pass（或者 noignore）：GDB 调试在捕获目标信号的同时，也允许程序自动处理该信号

比如，通过 `handle SIGINT nostop print pass` 告诉 GDB 在接收到 SIGINT 信号时不要停止，并把该信号传递给调试目标程序。

3.当 GDB 捕获到信号并暂停程序执行的那一刻，程序是捕获不到信号的，只有等到程序继续执行时，信号才能被程序捕获。

4.在 GDB 中使用 signal 函数手动给程序发送信号，比如`signal SIGINT`。

#### 调试多线程程序

1.常用命令：

|               命令                |                             功能                             |
| :-------------------------------: | :----------------------------------------------------------: |
|           info threads            | 查看当前调试环境中包含多少个线程，并打印出各个线程的相关信息，包括线程编号（ID）、线程名称等 |
|             thread id             |             将线程编号为 id 的线程设置为当前线程             |
|     thread apply id… command      | id… 表示线程的编号 list；command 代指 GDB 命令，如 next、continue 等。整个命令的功能是将 command 命令作用于指定编号的线程。如果想将 command 命令作用于所有线程，id… 可以用 all 代替 |
|     break location thread id      | 在 location 指定的位置建立普通断点，并且该断点仅用于暂停编号为 id 的线程 |
| set scheduler-locking off\on\step | 默认情况下，当程序中某一线程暂停执行时，所有执行的线程都会暂停；同样，当执行 continue 命令时，默认所有暂停的程序都会继续执行。该命令可以打破此默认设置，即只继续执行当前线程，其它线程仍停止执行 |

2.`set scheduler-locking` 中可选的调试模式：

- off：不锁定线程，即任何线程都可以随时执行
- on：锁定线程，只有当前线程或指定线程可以运行
- step：当用 `step` 命令单步调试某一线程时，其它线程不会执行，但是用其它命令（比如 `next`）调试线程时，其它线程也许会执行

3.对于多线程程序，GDB 默认采用的是 all-stop 模式，即只要有一个线程暂停执行，所有线程都暂停。这可以通过 `set non-stop on/off` 来打开/关闭 non-stop 模式。

4.non-stop 和 all-stop 模式的区别如下：

- non-stop 模式下可以在保持其它线程继续执行的状态下，单独调试某个线程
- 在 all-stop 模式下，`continue、next、step` 命令的作用对象是所有的线程；在 non-stop 模式下只作用于当前线程

#### 调试多进程程序

1.`attach pid`
利用该命令attach到子进程然后进行调试。为方便调试，可以sleep，这样有充分的时间进行调试。

2.在默认设置下，gdb调试多进程程序是只进入主（父）进程调试的，进入子进程则需要设置：
`set follow-fork-mode[parent | child]`

- parent：fork之后继续调试父进程，子进程不受影响
- child：fork之后调试子进程，父进程不受影响

`set detach-on-fork [on|off]`
detach-on-fork参数表明gdb在fork之后是否断开(detach)某个进程的调试，或者交给gdb控制。

- on：只调试父进程或子进程的其中一个(根据follow-fork-mode来决定)，这是默认的模式
- off：父子进程都在gdb的控制之下，其中一个进程正常调试(根据follow-fork-mode来决定)，另一个进程会被设置为暂停状态.

3.其他命令：

- 查看当前调试的fork进程的模式： `show follow-fork-mode`
- 查看detach-on-fork模式：`show detach-on-fork`
- 查询正在调试的进程：`info inferiors`
  gdb会为每个进程分配唯一的num，带*的即表示正在被调试的进程
- 切换调试的进程： `inferior num` (num为进程编号)
- `set schedule-multiple [on|off]`
  off：只有当前inferior会执行。
  on：全部是执行状态的inferior都会执行。
- `show schedule-multiple`，查看schedule-multiple的状态。
- `set follow-exec-mode [new|same]`
  same：当发生exec的时候，在执行exec的inferior上控制子进程。
  new：新建一个inferior给执行起来的子进程。而父进程的inferior仍然保留，当前保留的inferior的程序状态是没有执行。
- 查看follow-exec-mode设置的模式：`show follow-exec-mode`
- 打开和关闭inferior状态的提示信息 `set print inferior-events [on|off]`
- 查看print inferior-events设置的状态：`show print inferior-events`
- 显示当前gdb管理的地址空间的数目maint `info program-spaces`

#### 条件断点

1.条件断点就是满足某个条件才会触发的断点。
2.添加条件断点的命令是`break [lineNo] if [condition]`。
3.添加条件断点的另一个方法是先添加一个普通断点，然后使用命令`condition 断点编号 断点触发条件`。

#### 后台（异步）执行调试命令

在命令后面加上&，比如 `next&`

#### 反向调试

1.反向调试即时光倒流，指的是临时改变程序的执行方向，反向执行指定行数的代码，此过程中 GDB 调试器可以消除这些代码所做的工作，将调试环境还原到这些代码未执行前的状态。

2.常用命令：

|                命令                |                             功能                             |
| :--------------------------------: | :----------------------------------------------------------: |
|               record               | 让程序开始记录反向调试所必要的信息，其中包括保存程序每一步运行的结果等等信息。进行反向调试之前（启动程序之后），需执行此命令，否则是无法进行反向调试的 |
|        reverse-continue/rc         | 反向运行程序，直到遇到使程序中断的事件，比如断点或者已经退回到 record 命令开启时程序执行到的位置 |
|            reverse-step            | 反向执行一行代码，并在上一行代码的开头处暂停。和 step 命令类似，当反向遇到函数时，该命令会回退到函数内部，并在函数最后一行代码的开头处暂停执行 |
|            reverse-next            | 反向执行一行代码，并在上一行代码的开头处暂停。和 reverse-step 命令不同，该命令不会进入函数内部，而仅将被调用函数视为一行代码 |
|           reverse-finish           | 当在函数内部进行反向调试时，该命令可以回退到调用当前函数的代码处 |
| set exec-direction forward/reverse | forward 表示 GDB 以正常的方式执行所有命令；reverse 表示 GDB 将反向执行所有命令，由此我们可以直接只用step、next、continue、finish 命令来反向调试程序。注意，return 命令不能在 reverse 模式中使用 |

#### 调试脚本

1.我们通常都是在交互模式下使用 GDB 的，即手动输入各种 GDB 命令。其实 GDB 也支持执行预先写好的调试脚本，进行自动化的调试。调试脚本由一系列的 GDB 命令组成，GDB 会顺序执行调试脚本中的命令。

2.语法如下：

```shell
commands [breakpoint id]
… command-list …
end
```

当程序停在某个 breakpoint (或 watchpoint, catchpoint) 时（由 breakpoint id 标识），执行由 command-list 定义的一系列命令。
如果 commands 后不带参数，则默认标识的是最后一个断点。

3.建议 command-list 中第一个命令是 silent，这会让断点触发时打印的消息尽量精简。

4.command-list 中的最后一个命令通常是 `continue`，这样程序就不会在断点处停下，执行完一系列命令后可以继续执行。

5.写好自动化调试脚本后，运行该脚本的方式为：`gdb [program] -x [commands_file] > log`
举个栗子，有个名字是 test 的程序，我们想在 vfprintf 处打断点，打印每次触发到该断点时的调用堆栈：

```python
b vfprintf
commands 1
  silent
  bt
  c
end
c
```

将以上命令保存为脚本文件 auto.gdb，然后在终端执行 `gdb test -x auto.gdb >log`，这样就会把我们需要的信息打印到 log 文件中。

#### 查看命令手册

`help command` 可以查看 command 的描述、用法。

### 参考

- [100个gdb小技巧](https://wizardforcel.gitbooks.io/100-gdb-tips/content/index.html)
- [Debugging with GDB](https://sourceware.org/gdb/onlinedocs/gdb/index.html#SEC_Contents)
- [GDB 调试教程](http://c.biancheng.net/gdb/)
- [常用GDB命令速览](https://linux.cn/article-8900-1.html)
