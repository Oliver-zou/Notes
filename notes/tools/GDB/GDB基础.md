#### 一、调试准备

1.1 编译时加入调试信息：-g 选项
1.2 关闭编译器的程序优化选项。程序优化选项，一般有五个级别，从 O0 ~ O4，O0 表示不优化，从 O1 ~ O4 优化级别越来越高。
1.3 调试完之后可以移除调试信息：

- 编译时不加 -g 选项
- 使用 linux 的 strip 命令移除掉某个程序中的调试信息

#### 二、使用方式

##### 2.1 直接调试目标程序：`gdb filename`

##### 2.2 附加进程：`gdb -p pid`

当用 gdb attach 上目标进程后，调试器会暂停进程，此时可以使用 continue 命令让程序继续运行。当调试完程序想结束此次调试时，而且想让当前进程继续运行，则可以用 detach 命令让 GDB 调试器与程序分离。

##### 2.3  调试 core 文件：`gdb filename corename`

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

