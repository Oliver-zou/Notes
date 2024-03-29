[TOC]
# 简介
本文先介绍PMG模型作为基础. 以new一个对象为引子, 整体介绍Golang Runtime中内存结构和内存分配.
## 章节安排
本文整体分为三大部分.
第一部分介绍一点golang PMG调度模型的背景知识, 以及与内存分配的关系.
第二部分以new一个对象为引子, 简单介绍golang中的内存分配流程和策略.
第三部分深入解析go的内存结构和内存分配, 介绍mspan, mcache, mcentral, mheap等数据结构, 小于16Byte, 16Byte-32KB, >=32KB内存分配流程, 以及tiny分配器, fixalloc分配器, stack等几个特殊分配器.

# PMG调度模型简介

## 一、概述

**1.1 背景**

随着信息技术的迅速发展，单台服务器处理能力越来越强，迫使编程模式由从前的串行模式升级到并发模型。

并发模型包含 IO多路复用、多进程以及多线程，这几种模型都各有优劣，现代复杂的高并发架构大多是几种模型协同使用，不同场景应用不同模型，扬长避短，发挥服务器的最大性能。

**PS: **处理器访问任何寄存器和 Cache 等封装以外的数据资源都可以当成 I/O 操作，包括内存，磁盘，显卡等外部设备。）

而**多线程，因为其轻量和易用**，成为并发编程中使用频率最高的并发模型，包括后衍生的协程等其他子产品，也都基于它。

**1.2 并发 ≠ 并行**

并发 (concurrency) ：在单个 CPU 核上，线程通过时间片或者让出控制权来实现任务切换，达到 "同时" 运行多个任务的目的但实际上任何时刻都只有一个任务被执行，其他任务通过某种算法来排队。

并行 ( parallelism) ：多核 CPU 可以让同一进程内的 "多个线程" 做到真正意义上的同时运行，这才是并行。

**1.3 进程、线程、协程**

进程：进程是系统进行资源分配的基本单位，有独立的内存空间。

线程：线程是 CPU 调度和分派的基本单位，线程依附于进程存在，每个线程会共享父进程的资源。

协程：**协程是一种用户态的轻量级线程，**协程的调度完全由用户控制，协程间切换只需要保存任务的上下文，没有内核的开销。

**1.4 线程上下文切换（Thread Context Switch ）**

线程上下文是指某一时间点 CPU 寄存器和程序计数器的内容。

CPU执行线程的时候是通过时间分片的方式来轮流执行的，当某一个线程的时间片用完（到期），那么这个线程就会被中断，CPU不再执行当前线程，CPU会把使用权给其它线程来执行。如T1线程未执行结束，T2/T3线程插进来执行了，若干时间后T1又继续执行未执行完的部分，这种就造成了线程之间的来回切换。

一次上下文切换：CPU通过时间片分配算法来循环执行任务，当前任务执行一个时间片后会切换到下一个任务，在切换前会保存上一个任务的状态，以便下次切换回这个任务时，可以再次加载这个任务的状态，从任务保存到再加载的过程就是一次上下文切换。当Context Switch发生时，需要由操作系统保持当前线程的状态，并恢复另一个线程的状态，状态包括程序计数器、虚拟机栈中每个栈帧的信息。

造成原因：线程的CPU时间片用完；垃圾回收；有更高优先级的线程需要运行；线程自已调用了sleep、yield、wait、park、	synchronized、lock等方法

**1.5 上下文切换的开销**

上下文切换的开销包括直接开销和间接开销。

**上下文切换的代价是高昂的**，因为在核心上交换线程会花费很多时间。上下文切换的延迟取决于不同的因素，大概在在 50 到 100 纳秒之间。考虑到硬件平均在每个核心上每纳秒执行 12 条指令，那么一次上下文切换可能会花费 600 到 1200 条指令的延迟时间。实际上，上下文切换占用了大量程序执行指令的时间。直接开销有如下几点：

- 操作系统保存恢复上下文所需的开销
- 线程调度器调度线程的开销

如果存在**跨核上下文切换**（Cross-Core Context Switch），可能会导致 CPU 缓存失效（CPU 从缓存访问数据的成本大约 3 到 40 个时钟周期，从主存访问数据的成本大约 100 到 300 个时钟周期），这种场景的切换成本会更加昂贵。

间接开销有如下几点：

- 处理器高速缓存重新加载的开销
- 上下文切换可能导致整个一级高速缓存中的内容被冲刷，即被写入到下一级高速缓存或主存

多核CPU一定程度上可以减少上下文切换。

**1.6 线程调度模型**

1. 【1:1模型】一个用户线程和一个内核线程绑定，用户线程的调度依赖于内核线程的调度，这种方式可以充分利用多核资源，由cpu直接控制线程的调度。缺点是频繁的上下文切换和资源调度，如果用户线程较多，会带来很大的额外开销。

<div align="center"> <img src="../../pics/c5fe730e-2b53-4daa-ab57-1d6df7ec94be.png" width="500px"> </div><br>

​		优点：一个线程阻塞，其他线程并不会受到影响 缺点：创建一个用户线程就要创建		一个相应的内核线程。创建内核线程开销会影响应用程序的性能，所以这种模型的		大	多数实现限制了系统支持的线程数量。

2. 【N:1模型】多个用户线程绑定到一个内核线程，用户线程的调度切换可以用户自己实现，多个线程共享一个内核线程的时间片。这种方式极大减少了上下文切换，内核线程无需频繁切换调度，且不再受限于线程数，可以开辟大量用户线程。缺点是无法充分利用多核资源，对于多个用户线程而言，其本质仍然是单核运行，需要分时间片利用内核线程的cpu时间。

   <div align="center"> <img src="../../pics/bb82c644-0cb9-4207-baff-df3e7f03b924.png" width="500px"> </div><br>

   优点：高效的上下文切换，几乎无影响的线程数量 缺点：一个线程阻塞，所有线程无法执行，多核CPU处理器上对性能不会有显著提高。	

3. 【N:M模型】用户线程和内核线程是多对多的关系，每个用户线程可以被多个内核线程调度，每个内核线程可以调度多个不同的用户线程。这种调度模型综合了前面两种的优缺点。既可以充分利用多核资源，每个空闲的内核线程都可以参与调度，也可以减少上下文开销，每个内核线程可以在一定时间内运行多个用户线程。但是这种调度模型实现复杂度高。go的调度是基于这种模式。

   <div align="center"> <img src="../../pics/d4f6a64b-2647-4780-91c6-aed4ff25290e.png" width="500px"> </div><br>
   
   避免了两个缺点，开发人员可以创建任意多的用户线程，并且相应的内核线程能在多处理器上并发执行，而且当一个线程调度的时候，内核可以调度另一个线程来执行。

**1.7 go 协程的特点**

用户态线程又可以被称作为协程。协程状态：等待、可运行、执行 Go协程（有栈协程）只存在go运行时的虚拟空间，go运行时调度器着goroutines的生命周期。Go运行时维护着三个C语言结构： 1.G结构体：代表着独立的go routine，包含栈指针，ID、缓存以及状态属性 2.M结构体：代表一个操作系统，他包含一个指向包含可执行goroutines全局队列的指针,当前正在执行的goroutines以及调度器的引用 3.Sched结构体：全局结构体，包含空闲队列，等待协程队列，线程也有相应的队列。 所以，在启动的时候，go运行时其启动一些协程来进行gc，调度以及完成用户代码，实际在操作系统层面，只有一个线程来处理这些协程。

## 二、Golang的出现

### **Golang为并发而生**

Golang从2009年正式发布以来，依靠其极高运行速度和高效的开发效率，迅速占据市场份额。Golang从语言级别支持并发，通过轻量级协程Goroutine来实现程序并发运行。 

**Goroutine**非常轻量，主要体现在以下两个方面：

**上下文切换代价小：** Goroutine上下文切换只涉及到三个寄存器（PC / SP / DX）的值修改；而对比线程的上下文切换则需要涉及模式切换（从用户态切换到内核态）、以及16个寄存器、PC、SP...等寄存器的刷新；

**内存占用少：**线程栈空间通常是2M，Goroutine栈空间最小2K；

Golang程序中可以轻松支持**10w****级别**的Goroutine运行，而线程数量达到1k时，内存占用就已经达到2G。

### **Go调度器**实现机制：

Go程序通过调度器来调度**Goroutine****在内核线程上执行，**但是G - *Goroutine*并不直接绑定OS线程M - *Machine*运行，而是由Goroutine Scheduler中的 P - *Processor* （逻辑处理器）来作获取内核线程资源的『中介』。

Go调度器模型我们通常叫做**G-P-M**模型**，他包括4个重要结构，分别是**G**、P、M、**Sched：

**G:Goroutine**，每个Goroutine对应一个G结构体，G存储Goroutine的运行堆栈、状态以及任务函数，可重用。

G并非执行体，每个G需要绑定到P才能被调度执行。

**P: Processor**，表示逻辑处理器，对G来说，P相当于CPU核，G只有绑定到P才能被调度。对M来说，P提供了相关的执行环境(Context)，如内存分配状态(mcache)，任务队列(G)等。

P的数量决定了系统内最大可并行的G的数量（前提：物理CPU核数 >= P的数量）。

**P的数量由用户设置的GoMAXPROCS决定，但是不论GoMAXPROCS设置为多大，P的数量最大为256。**

**M: Machine**，OS内核线程抽象，代表着真正执行计算的资源，在绑定有效的P后，进入schedule循环；而schedule循环的机制大致是从Global队列、P的Local队列以及wait队列中获取。

**M的数量是不定的，由Go Runtime调整，**为了防止创建过多OS线程导致系统调度不过来，目前默认最大限制为10000个。

M并不保留G状态，这是G可以跨M调度的基础。

**Sched**：Go调度器，**它维护有存储M和G的队列以及调度器的一些状态信息等。

调度器循环的机制大致是从各种队列、P的本地队列中获取G，切换到G的执行栈上并执行G的函数，调用Goexit做清理工作并回到M，如此反复。

**理解M、P、G三者的关系，可以通过经典的地鼠推车搬砖的模型来说明其三者关系：**

<div align="center"> <img src="../../pics/bbd34ade-5c0a-4c81-beab-78de3c062d62.png" width="500px"> </div><br>

**地鼠(Gopher)的工作任务是：**工地上有若干砖头，地鼠**借助小车**把砖头运送到火种上去烧制。**M**就可以看作图中的地鼠，P就是小车，G就是小车里装的砖。

弄清楚了它们三者的关系，下面我们就开始重点聊地鼠是如何在搬运砖块的。

**Processor（P）：**

根据用户设置的 **GoMAXPROCS** 值来创建一批小车(P)。

**Goroutine(G)**：

通过Go关键字就是用来创建一个 Goroutine，也就相当于制造一块砖(G)，然后将这块砖(G)放入当前这辆小车(P)中。

**Machine (M)**：

地鼠(M)不能通过外部创建出来，只能砖(G)太多了，地鼠(M)又太少了，实在忙不过来，**刚好还有空闲的小车(P)没有使用**，那就从别处再借些地鼠(M)过来直到把小车(P)用完为止。

这里有一个地鼠(M)不够用，从别处借地鼠(M)的过程，这个过程就是创建一个内核线程(M)。

**需要注意的是：**地鼠(M) 如果没有小车(P)是没办法运砖的，**小车(P)的数量决定了能够干活的地鼠(M)数量**，在Go程序里面对应的是活动线程数；

**在Go程序里我们通过下面的图示来展示G-P-M模型：**

<div align="center"> <img src="../../pics/1582289502_14_w554_h259.png" width="500px"> </div><br>

P代表可以“并行”运行的逻辑处理器，每个P都被分配到一个系统线程M，G 代表 Go 协程。

Go 调度器中有两个不同的运行队列：**全局运行队列(GRQ)和本地运行队列(LRQ)。**

每个P都有一个LRQ，用于管理分配给在P的上下文中执行的 Goroutines，这些 Goroutine 轮流被和P绑定的M进行上下文切换。GRQ 适用于尚未分配给P的 Goroutines。

**从上图可以看出，G的数量可以远远大于M的数量，换句话说，Go程序可以利用少量的内核级线程来支撑大量Goroutine的并发。**多个Goroutine通过用户级别的上下文切换来共享内核线程M的计算资源，但对于操作系统来说并没有线程上下文切换产生的性能损耗。

**为了更加充分利用线程的计算资源，Go调度器采取了以下几种调度策略：**

**任务窃取（work-stealing）**

我们知道，现实情况有的Goroutine运行的快，有的慢，那么势必肯定会带来的问题就是，忙的忙死，闲的闲死，Go肯定不允许摸鱼的P存在，势必要充分利用好计算资源。

为了提高Go并行处理能力，调高整体处理效率，当每个P之间的G任务不均衡时，调度器允许从GRQ，或者其他P的LRQ中获取G执行。

**减少阻塞**

如果正在执行的Goroutine阻塞了线程M怎么办？P上LRQ中的Goroutine会获取不到调度么？

**在Go里面阻塞主要分为一下4种场景：**

**场景1：由于原子、互斥量或通道操作调用导致 Goroutine 阻塞**，调度器将把当前阻塞的Goroutine切换出去，重新调度LRQ上的其他Goroutine；

**场景2**：**由于网络请求和IO操作导致 Goroutine 阻塞**，这种阻塞的情况下，我们的G和M又会怎么做呢？

Go程序提供了**网络轮询器（NetPoller）**来处理网络请求和IO操作的问题，其后台通过kqueue（MacOS），epoll（Linux）或 iocp（Windows）来实现IO多路复用。

通过使用NetPoller进行网络系统调用，调度器可以防止 Goroutine 在进行这些系统调用时阻塞M。这可以让M执行P的 LRQ 中其他的 Goroutines，而不需要创建新的M。有助于减少操作系统上的调度负载。

**下图展示它的工作原理：**G1正在M上执行，还有 3 个 Goroutine 在 LRQ 上等待执行。网络轮询器空闲着，什么都没干。

<div align="center"> <img src="../../pics/1582289670_35_w281_h147.png" width="500px"> </div><br>

接下来，G1想要进行网络系统调用，因此它被移动到网络轮询器并且处理异步网络系统调用。然后，M可以从 LRQ 执行另外的 Goroutine。此时，G2就被上下文切换到M上了。

<div align="center"> <img src="../../pics/1582289723_54_w290_h154.png" width="500px"> </div><br>

最后，异步网络系统调用由网络轮询器完成，G1被移回到P的 LRQ 中。一旦G1可以在M上进行上下文切换，它负责的 Go 相关代码就可以再次执行。这里的最大优势是，执行网络系统调用不需要额外的M。网络轮询器使用系统线程，它时刻处理一个有效的事件循环。

<div align="center"> <img src="../../pics/1582289748_28_w303_h161.png" width="500px"> </div><br>

这种调用方式看起来很复杂，值得庆幸的是，**Go语言将该“复杂性”隐藏在Runtime中**：Go开发者无需关注socket是否是 non-block的，也无需亲自注册文件描述符的回调，只需在每个连接对应的Goroutine中以“block I/O”的方式对待socket处理即可，**实现了goroutine-per-connection简单的网络编程模式**（但是大量的Goroutine也会带来额外的问题，比如栈内存增加和调度器负担加重）。

用户层眼中看到的Goroutine中的“block socket”，实际上是通过Go runtime中的netpoller通过Non-block socket + I/O多路复用机制“模拟”出来的。Go中的net库正是按照这方式实现的。

**场景3**：当调用一些系统方法的时候，如果系统方法调用的时候发生阻塞，这种情况下，网络轮询器（NetPoller）无法使用，而进行系统调用的 Goroutine 将阻塞当前M。

让我们来看看同步系统调用（如文件I/O）会导致M阻塞的情况：G1将进行同步系统调用以阻塞M1。

<div align="center"> <img src="../../pics/1582289777_96_w321_h190.png" width="500px"> </div><br>

调度器介入后：识别出G1已导致M1阻塞，此时，调度器将M1与P分离，同时也将G1带走。然后调度器引入新的M2来服务P。此时，可以从 LRQ 中选择G2并在M2上进行上下文切换。

<div align="center"> <img src="../../pics/1582289815_93_w347_h163.png" width="500px"> </div><br>

阻塞的系统调用完成后：G1可以移回 LRQ 并再次由P执行。如果这种情况再次发生，M1将被放在旁边以备将来重复使用。

<div align="center"> <img src="../../pics/1582682239_99_w363_h171.png" width="500px"> </div><br>

**场景4**：如果在Goroutine去执行一个sleep操作，导致M被阻塞了。

Go程序后台有一个监控线程sysmon，它监控那些长时间运行的G任务然后设置可以抢占的标识符，别的Goroutine就可以抢先进来执行。

只要下次这个Goroutine进行函数调用，那么就会被强占，同时也会保护现场，然后重新放入P的本地队列里面等待下次执行。





## 三、Goroutine调度模型 

## PMG简介
这里简单介绍一下Go的调度模型, 这个调度模型与Go的内存分配结构有很大的联系.
G是协程,  非常轻量级,  初始化栈大小仅2KB(不同系统不同).
P代表一个逻辑Processor, 是资源的拥有者, 包含了per-P的内存分配资源, 本地运行队列等, P的个数限制最大可同时运行的G的个数, 初始化时设置为核的个数或者自定义的GOMAXPROCS变量. per-P数据结构的访问是无锁的.
M则对应了一个系统线程, 如果要运行Go代码(而不是系统调用或者cgo), M需要与一个P结合, 通过调度寻找可运行的G, 真正执行G中的func.
![PMG调度图](http://km.oa.com/files/photos/pictures//20181127//1543322454_59.png)

如上图所示，每个M都与一个内核线程绑定，在go的运行过程中，M的绑定关系不变。每个M在同一时刻都至多只能与一个P绑定，每个P都有一个自己的本地队列，M通过从P的本地队列中取一个G执行，如果本地队列消耗完毕，则会从全局队列取一个G，如果全局队列也消耗完毕，则从其他P那里窃取G。

**情况1.G阻塞**

如果一个被调度的G（G1）进入阻塞态，此时M将无法继续执行P中的其他G。如果G1因系统调用被阻塞，M会和P解绑，P会被另一个M绑定，当前M等待系统调用返回，然后尝试获取一个P，如果获取成功，则将G1加入该P队列，如果获取失败，则将G1放入全局队列，M自己放入空闲M等待队列，等待一个可绑定的P。如果G1是因为IO、管道等用户态操作阻塞，则G1会被放入等待队列，当前M继续执行P中其他G，当G1被另一个G唤醒的时候，再加入对应P。

**情况2.P中G消耗完毕**

不同P中G不一样，执行效率也不一样，导致可能一个P很快就消耗完了本地队列，全局队列也无新的G，此时为了任务均衡，P会尝试从其他P中窃取一半的G加入自己队列。

**情况3.全局队列中有G，但是所有P都能生产消费平衡。**

如果每个M都只执行对应P中本地队列的G，只有当P中无G才去消费全局队列，那么在P的本地队列一直不为空的情况下，全局队列中的G将一直无法被调度到。所以每次M都有一定概率从全局队列中找G，以保证调度公平。

## 为什么要有PMG

**为什么要有PMG模型? MG不就够了吗? PMG和内存有什么联系?**

<div align="center"> <img src="../../pics/1634232792-5927-616869d890ba8-655683.png" width="500px"> </div><br>

在 `Go 1.1`版本之前，其实用的就是`GM`模型。

- **G**，协程。通常在代码里用 `go` 关键字执行一个方法，那么就等于起了一个`G`。
- **M**，**内核**线程，操作系统内核其实看不见`G`和`P`，只知道自己在执行一个线程。`G`和`P`都是在**用户层**上的实现。

除了`G`和`M`以外，还有一个**全局协程队列**，这个全局队列里放的是多个处于**可运行状态**的`G`。`M`如果想要获取`G`，就需要访问一个**全局队列**。同时，内核线程`M`是可以同时存在多个的，因此访问时还需要考虑**并发**安全问题。因此这个全局队列有一把**全局的大锁**，每次访问都需要去获取这把大锁。

并发量小的时候还好，当并发量大了，这把大锁，就成为了**性能瓶颈**。

存在以下问题:

1. 调度锁问题. 单一的全局调度锁(Sched.Lock)和集中的状态, 导致伸缩性下降. 所有和goroutine相关比如创建, 完成, 重新调度等操作都要争抢这把锁.

   即所有可运行G全部放在全局队列中，M从全局队列中获取可运行的G，多线程访问同一资源需要加锁进行互斥/同步。所以全部G队列时互斥锁进行保护的。在Vtocc Server 8核，CPU达到70%的，发现14%的都消耗在锁上。

2. G传递问题. 在工作线程M之间需要经常传递runnable的G, 这个会加大调度延迟, 并带来额外的性能损耗.比如当G中包含创建新的协程的时候，M创建了G2，但是为了继续执行G，需要其他M来执行G2，造成的很差的局部性。

3. **Per-M的内存问题. 基于TCMalloc结构的内存结构, 每个M都需要memory cache和其他类型的cache(stack alloc), 然而实际上只有M在运行Go代码时才需要这些Per-M Cache, 阻塞在系统调用的M并不需要mcache. 正在运行Go代码的M与进行系统调用的M的比例可能高达1:100, 这造成了极大的内存消耗, 和很差的数据局部性**

4. 线程阻塞与从阻塞恢复. 由于系统调用而形成的剧烈的worker thread阻塞和解除阻塞，导致很大的性能损耗.

Dmitry Vyukov提出和实现了PMG模型(Go 1.2)
Scalable Go Scheduler Design Doc
https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw/edit

基于**没有什么是加一个中间层不能解决的**思路，`golang`在原有的`GM`模型的基础上加入了一个调度器`P`，可以简单理解为是在`G`和`M`中间加了个中间层。

于是就有了现在的`GMP`模型里。

- `P` 的加入，还带来了一个**本地协程队列**，跟前面提到的**全局队列**类似，也是用于存放`G`，想要获取等待运行的`G`，会**优先**从本地队列里拿，访问本地队列无需加锁。而全局协程队列依然是存在的，但是功能被弱化，不到**万不得已**是不会去全局队列里拿`G`的。
- `GM`模型里M想要运行`G`，直接去全局队列里拿就行了；`GMP`模型里，`M`想要运行`G`，就得先获取`P`，然后从 `P` 的本地队列获取 `G`。

- 新建 `G` 时，新`G`会优先加入到 `P` 的本地队列；如果本地队列满了，则会把本地队列中一半的 `G` 移动到全局队列。
- `P` 的本地队列为空时，就从全局队列里去取。
- 如果全局队列为空时，`M` 会从其他 `P` 的本地队列**偷（stealing）一半G**放到自己 `P` 的本地队列。
- `M` 运行 `G`，`G` 执行之后，`M` 会从 `P` 获取下一个 `G`，不断重复下去。

**通过P解耦G和M, 将之前与M关联的mcache移至P中. 极大的减少了内存消耗.**

## **为什么P的逻辑不直接加在M上**

主要还是因为`M`其实是**内核**线程，内核只知道自己在跑线程，而`golang`的运行时（包括调度，垃圾回收等）其实都是**用户空间**里的逻辑。操作系统内核哪里还知道，也不需要知道用户空间的golang应用原来还有那么多花花肠子。这一切逻辑交给应用层自己去做就好，毕竟改内核线程的逻辑也不合适啊。

# 从new说起
## new一个对象
没有对象, 就new一个. 在golang里分配一个对象的方式有以下几种:

1.new(TypeA)
2.&TypeA{}
3.make

其中1和2是一样, 2和3最终都会调同一个方法. 1和3是语言的关键字, 2是字面量定义.
## 逃逸分析
我们都知道对于有GC的语言, 大部分对象都是在语言运行时的堆中分配的.
但是并不是每次new一个对象都会在go的堆上分配, 如果语言编译时能够检测出, 这个对象比较小, 且生命周期不会脱离当前栈的范围, 那么其实是可以把这个对象分配在栈上面的. 这就叫做逃逸分析,特别是对于存在即意味着消亡的小对象来说, 优化明显, 加快了内存分配速度, 减少了需要GC回收的对象数量, 减少了GC压力.Java, Go这些带GC的语言都有类似的功能. 这里不过多展开.

## new为何物?
前面讲到分配一块内存的几种方式, new, make关键字和字面量. 这些其实都是语法糖, 最终还是会像我们在C里面一样调用方法来分配对象.
我们看一段代码:
```go
package main
import "fmt"
func main() {
	fmt.Printf("%d", *testFunc())
}
func testFunc() *int {
	return new(int)
}
```
```bash
go build -gcflags="-l -N" # -l 禁止内联, -N 禁止优化, go tool compile --help查看更多编译参数的帮助
go  tool objdump -s "main\.testFunc" -S ./newTest # -s 查看符合该正则表示的汇编符号信息, -S打印出源码
```
输出结果如下:
```
TEXT main.testFunc(SB) /home/yifhao/go/src/go_learning/src/plan9/newTest.go
func testFunc() *int {
....
        return new(int)
  0x483a46              488d0573f80000          LEAQ 0xf873(IP), AX
  0x483a4d              48890424                MOVQ AX, 0(SP)
  0x483a51              e8caa3f8ff              CALL runtime.newobject(SB)
....

```
在runtime/malloc.go文件中, newobject方法如下.
```go
//new的实现, _type是编译器进行生成的
func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}

//new调用的函数, 除了new以外, make关键字编译出来的对应函数也会调用这个函数
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	...
}
```

对于make一个slice或者make一个map.
```go
//runtime/slice.go
//make一个slice, 编译器生成以下方法.
func makeslice(et *_type, len, cap int) slice {
	maxElements := maxSliceCap(et.size)
	if len < 0 || uintptr(len) > maxElements {
		panicmakeslicelen()
	}

	if cap < len || uintptr(cap) > maxElements {
		panicmakeslicecap()
	}

	p := mallocgc(et.size*uintptr(cap), et, true)
	return slice{p, len, cap}
}

//runtime/map.go
//make里面是一个map, 稍微复杂一些, 不过最终也是调用new和mallocgc
func makemap(t *maptype, hint int, h *hmap) *hmap {
....
}
```

## 万物归于mallocgc
go里面new一个对象, 最终编译器生成的是func newobject(typ \*\_type) unsafe.Pointer方法, 里面在调用mallocgc方法. 而make也会调用mallocgc方法. mallocgc是我们日常写go代码中申请内存的入口.

## 流程代码
我们来简单看一下mallocgc的代码结构.
```go
//runtime/malloc.go
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	//如果size为0, 返回一个全局对象的地址. 这也就是用channel传递信号时, 建议用空struct{}的原因, 所有的空结构体类型的实例, 都是一个.
	if size == 0 {
		return unsafe.Pointer(&zerobase)
	}

	dataSize := size
	c := gomcache()
	var x unsafe.Pointer
	//要分配的对象的上有没有指针, 没有指针意味着这个对象在gc时不需要被扫描.
	noscan := typ == nil || typ.kind&kindNoPointers != 0
	//maxSmallSize,32KB. 如果对象小于32KB.
	if size <= maxSmallSize {
		//对象上没有指针, 且小于16字节, 使用tiny分配器. tiny分配器是分配小对象的一个优化, 一小块空间(16字节), 以指针移动的方式进行分配.
		if noscan && size < maxTinySize {
			//上一次分配到的位置.
			off := c.tinyoffset
			//根据大小进行对齐
			if size&7 == 0 {
				off = round(off, 8)
			} else if size&3 == 0 {
				off = round(off, 4)
			} else if size&1 == 0 {
				off = round(off, 2)
			}
			//maxTinySize为16字节, 每个tiny分配器大小为16字节.用完了再去取一个.
			//如果还有空间, 就使用当前tiny分配器分配
			if off+size <= maxTinySize && c.tiny != 0 {
				// The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				//分配结束, 返回对象指针
				return x
			}
			//如果当前的tiny分配器不够用了, 那就再分配一个tiny分配器咯
			// Allocate a new maxTinySize block.
			span := c.alloc[tinySpanClass]
			//获取到了一个新的tiny分配器
			v := nextFreeFast(span)
			if v == 0 {
				v, _, shouldhelpgc = c.nextFree(tinySpanClass)
			}
			//分配的对象地址
			x = unsafe.Pointer(v)
			//因为最大只有16字节, 就简单的初始化一下0
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			//小优化, 如果分配完当前对象, 新的tiny分配器剩余的字节数比当前的多, 就替换对象
			if size < c.tinyoffset || c.tiny == 0 {
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
			//这里没有直接返回, 因为重新分配了一个tiny分配器, 后面有一些事情要处理, 暂时不关心

		} else { //如果对象小于32KB, 且(大于16个字节或者对象中包含指针)
			var sizeclass uint8
			//获取该对象大小对应的大小级别.
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
			} else {
				sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
			}
			size = uintptr(class_to_size[sizeclass])
			spc := makeSpanClass(sizeclass, noscan)
			//获取对象大小对应的span
			span := c.alloc[spc]
			//分配对象
			v := nextFreeFast(span)
			if v == 0 {
				v, span, shouldhelpgc = c.nextFree(spc)
			}
			x = unsafe.Pointer(v)
			...
		}
	} else { //对象>=32KB, 则直接从heap中分配.
		var s *mspan
		shouldhelpgc = true
		systemstack(func() {
			s = largeAlloc(size, needzero, noscan)
		})
		s.freeindex = 1
		s.allocCount = 1
		x = unsafe.Pointer(s.base())
		size = s.elemsize
	}
	..... //省略一些代码

	//返回对象地址
	return x
}
```
## 内存分配过程
从上面的代码可以看出来, go代码中分配内存的过程大致如下:
1. 对于size为0的对象, 比如一个空的strcut{}, 分配器做了优化, 直接返回一个特殊的全局变量zerobase.
2. 小于16字节的无指针(noscan)的逃脱的小对象, 比如new(int), 这些直接在每个P上的tiny分配器上分配, 简单的说就是移动指针的方式分配, 小块内存可以分配多个小对象.下面会有详细介绍.
3. 如果对象大于16字节小于32K, 则从与当前P绑定的cache中对应的class的span里分配(无锁). 对象中有指针从分配有指针对象的span中分配, 没有指针的从noscan的span分配. 如果当前P对应的span用完了, 那么从该class对应的后备mcentral中分配一个span, 如果mcentral都没有了, 则从mheap中获取, mheap没有空闲的, 则mmap从系统获取新的内存. 一步步往上升, 从per-P cache的无锁到每个class对应的mcentral(有锁, 但有67\*2个, 粒度较小), 再到全局锁的mheap. 大部分分配在per-P的mcache就可以解决, 大大减少了全局竞争的可能.
4. 如果对象大于32K, 则调用largeAlloc. largeAlloc先从heap中维护的类似于伙伴分配器的地方分配, 如果没有多余的连续空闲页, 则使用mmap从系统分配.

# 分割线
通过前面我们大致可以知道golang内存分配大致流程.

总体上是类似于TCMalloc的结构, 同时golang在内存分配上做了很多优化.

1. 不同大小的内存需求使用不同class的span进行分配, 减少内存碎片.
1. 每个P都带有不同class的span, 用于无锁分配.
1. heap里采用了类似伙伴分配器的方式管理多余的页, 进行页分类和合并.
1. 0size的对象, 共用一个全局对象, 对于小对象, 采用tiny分配器分配,对于16字节-32KB的对象, 采用不同span分配, 对于>=32KB的对象, 直接从heap里分配.

如果只是简单的了解go内存结构和分配策略, 看到这里就可以了, 但是我知道大家肯定忍不住好奇心, 大家系好安全带, 起飞!
下面章节我将结合源码和图片来深入讲解go的内存模型.

# Golang Runtime内存结构
Go程序在启动时会进行多个步骤的初始化, call osinit, call schedinit, make & queue new G, call runtime·mstart等. 其中schedinit步骤会进行stackinit(), mallocinit(), gcinit()等步骤. 在mallocinit步骤中, 会计算并初始化Golang虚拟内存布局. 在linux amd64位机器上, 会计算出以下的虚拟内存结构(注意只是计算, arena区域并不实际分配这么大的虚拟空间)
![golang amd64内存分区结构](http://km.oa.com/files/photos/pictures//20181201//1543652015_39.png)
## arena, spans, bitmap功能描述
1. arena为后续Go程序分配对象的地址空间, 也就是之前mallocgc方法涉及的内存, 也叫go gc heap, linux amd64上假定大小为512G,  为了方便管理把arena区域划分成一个个的page,  每个page 8KB,  一共有512GB/8KB个页, 每个page按照顺序编号, arena_start至arena_start+8KB为page 1, 接下来8KB为page 2, 以此类推.

2. spans区域存放指向span的指针,  表示arean中对应的Page所属的span, 所以spans区域的大小为(512GB/8KB)\*指针大小8byte = 512M. 比如分配了一个2个页的span A(地址为&A), 使用了page 10, page 11, 那么spans区域spans_start+8\*9和spans_start+8\*10的内容都为span A的地址&A. 以此来建立任意地址到span的这样一个索引, span本身就已经有了span到其所拥有的地址(page start的index和end的index)的索引, 这样就建立了两者的双向索引.

3. bitmap是用于辅助GC的, arena中的每个字(8字节)都在bitmap中有两个bit表示其状态, 该word是否是指针(func (h heapBits) isPointer()),是否应该继续扫描(该字所属的object后面位置是否还有指针,func (h heapBits) morePointers() bool ).

## Runtime内存结构分配代码及运行时情况
分配代码如下:
```go
//malloc.go mallocinit函数中. 初始化化虚拟内存布局
		arenaSize := round(_MaxMem, _PageSize)
		pSize = bitmapSize + spansSize + arenaSize + _PageSize
		for i := 0; i <= 0x7f; i++ {
			switch {
			case GOARCH == "arm64" && GOOS == "darwin":
				p = uintptr(i)<<40 | uintptrMask&(0x0013<<28)
			case GOARCH == "arm64":
				p = uintptr(i)<<40 | uintptrMask&(0x0040<<32)
			default:
				p = uintptr(i)<<40 | uintptrMask&(0x00c0<<32)
			}
			//p为选择的基地址, pSize为上面spans, bitmap, 加上arena的大小. 
			//linux amd64位中, 并不是真正一开始就分配这么多虚拟内存. 
			/对于大于4GB的内存分配, 只是试探性的mmap分配64KB,然后munmap掉. 以表示这段内存是可用的.
			p = uintptr(sysReserve(unsafe.Pointer(p), pSize, &reserved))
			if p != 0 {
				break
			}
		}
```

```go
// mem_linux.go尝试分配的代码.
func sysReserve(v unsafe.Pointer, n uintptr, reserved *bool) unsafe.Pointer {
	// On 64-bit, people with ulimit -v set complain if we reserve too
	// much address space. Instead, assume that the reservation is okay
	// if we can reserve at least 64K and check the assumption in SysMap.
	// Only user-mode Linux (UML) rejects these requests.
	//翻译一下, 应该是之前大的虚拟地址也是直接保留的.
	//后来用户说占了太大的虚拟内存, 就改了, 先分配64K看看, 如果是ok的, 那么认为保留那块大的内存也是ok的.
	if sys.PtrSize == 8 && uint64(n) > 1<<32 {
		//当初始化内存结构时, 是走入这个分支. 使用mmap, flag中加fix.
		p, err := mmap_fixed(v, 64<<10, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
		if p != v || err != 0 {
			if err == 0 {
				munmap(p, 64<<10)
			}
			return nil
		}
		munmap(p, 64<<10)
		*reserved = false
		return v
	}

	p, err := mmap(v, n, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		return nil
	}
	*reserved = true
	return p
}
```
![程序运行时的情况](http://km.oa.com/files/photos/pictures//20181201//1543653921_36.png)
# 几个重要的结构体
接下来会讲解Golang Runtime代码中比较重要的结构体.
**mspan**是管理分配不同大小class段的内存块的结构体, 一共有67\*2个class.
每个P都拥有一个自己的**mcache**,  对其操作时不需要锁.每个mcache有67\*2个mspan,  和一个tiny分配器.
mcache中每个class的mspan都对应一个后备的全局的**mcentral** ,用于分配对应class的mspan给mcache. 所以mcentral全局也有67\*2个.
**mheap**全局一个, 包含不同page大小的空闲块的page list(free), 超过128个连续空闲页的块(freelarge)的list, 当然也有对应的已经分配出去的页的list(busy, busylarge), 所有span的slice, spans区域的内存映射的slice, bitmap的地址, arena状态相关的地址, mcentral的引用等.

以上结构体大部分功能是用分配go的用户代码需要的内存, 即arean这段内存, 当然这些结构体本身并不是在arean中分配.(后面会专门讲一讲fixalloc). 整体结构图大致如下.
![整体图](http://km.oa.com/files/photos/pictures//20190303//1551614469_82.png)
# mspan
前面讲到golang为解决分配对象时的内存浪费问题, 采用了span的方式.
## span级别, span对应分配对象大小, span大小, 内存浪费率
类似于linux的slab分配器和TCMalloc. 不同大小的对象使用不同class的span分配.

go中的span一共有67个size. 而有的对象的字段中有指针, 有的没指针, 没有指针的对象在GC扫描时不需要继续扫描, 因为它不引用其他对象, 所以每个size又分了scan和noscan. 所以go中一共有(66+1) * 2=134个span类型. (每个大于32KB的分配请求, 都使用一个class0的span).

```go
//runtime/sizeclasses.go
const _NumSizeClasses = 67
var class_to_size = [_NumSizeClasses]uint16{0, 8, 16, 32, 48, 64, 80, 96, 112, 128, 144, 160, 176, 192, 208,
224, 240, 256, 288, 320, 352, 384, 416, 448, 480, 512, 576, 640, 704, 768, 896, 1024, 1152, 1280, 1408, 1536,
1792, 2048, 2304, 2688, 3072, 3200, 3456, 4096, 4864, 5376, 6144, 6528, 6784, 6912, 8192, 9472, 9728, 10240,
10880, 12288, 13568, 14336, 16384, 18432, 19072, 20480, 21760, 24576, 27264, 28672, 32768}
```

以class5为例, class用于分配49-64(包含)字节的对象, 这样的一个span拥有一个8K的页(go中一个page为8K), 每个span最多可以分配128个对象, 最小内存浪费率为0, 最大为23.44%, 即都分配了49bytes的对象, `1-(49*128)/8192=0.234375`.

以下来源于runtime/sizeclasses.go文件中的注释.

```shell
// class  bytes/obj  bytes/span  objects  tail waste  max waste
//     1          8        8192     1024           0     87.50%
//     2         16        8192      512           0     43.75%
//     3         32        8192      256           0     46.88%
//     4         48        8192      170          32     31.52%
//     5         64        8192      128           0     23.44%
//     6         80        8192      102          32     19.07%
//     7         96        8192       85          32     15.95%
//     8        112        8192       73          16     13.56%
//     9        128        8192       64           0     11.72%
//    10        144        8192       56         128     11.82%
//    11        160        8192       51          32      9.73%
//    12        176        8192       46          96      9.59%
//    13        192        8192       42         128      9.25%
//    14        208        8192       39          80      8.12%
//    15        224        8192       36         128      8.15%
//    16        240        8192       34          32      6.62%
//    17        256        8192       32           0      5.86%
//    18        288        8192       28         128     12.16%
//    19        320        8192       25         192     11.80%
//    20        352        8192       23          96      9.88%
//    21        384        8192       21         128      9.51%
//    22        416        8192       19         288     10.71%
//    23        448        8192       18         128      8.37%
//....省略一部分
//    61      20480       40960        2           0      6.87%
//    62      21760       65536        3         256      6.25%
//    63      24576       24576        1           0     11.45%
//    64      27264       81920        3         128     10.00%
//    65      28672       57344        2           0      4.91%
//    66      32768       32768        1           0     12.50%
```

## mspan图示
每个mspan有指向1个或多个页的内存块的指针, 被切分成多个slot. 每个mspan代表用于分配同一个class内存的块,
![mspan](http://km.oa.com/files/photos/pictures//20190224//1551012795_4.png)
## 详解mspan结构体
golang中用于分配对象的mspan分配对象的时候.
使用freeindex配合allocCache来分配对象.
freeindex表示0-freeindex之间的slot是确定已经被分配的, 要想寻找没有被分配的slot, 结合allocCache来查找.
allocBits来标记上一次GC的sweep阶段之后, 哪些slot对应的对象仍然是被使用的.即为0的位置就是空闲的.
而allocCache就是allocBits在freeindex位之后的64个位的数据.
每次freeindex都会增加, 而allocCache则会移位.
![寻找方式](http://km.oa.com/files/photos/pictures//20190303//1551615267_61.png)
```
做了适当精简.
type mspan struct {
  //mspan在mcentral或者mheap时构成的链表
	next *mspan     // next span in list, or nil if none
	prev *mspan     // previous span in list, or nil if none
	list *mSpanList // For debugging. TODO: Remove.

   //mspan的开始地址, 是至少8KB对齐
	startAddr uintptr // address of first byte of span aka s.base()
	//占用的页表(golang中目前定义的page为8KB)
	npages    uintptr // number of pages in span


	//表示从这一个偏移开始寻找free的对象, 0-freeindex之间的对象都分配完了. 如果freeindex等于nelems的话, 那么这个span就全分配完了.
	//这个字段和下面的allocBits一起使用. 如果某个位置n>=freeindex, 且allocBits[n/8]=0的话, 那么这个位置就是未分配的. 否则就是已经分配的.
	//第n个对象在内存中的起始位置在 startAddr+n*elemsize, 其中start为这个span开始地址, elemsize为这个span用于分配的对象的大小.
	freeindex uintptr

	//这个mspan被划分成多少个对象区. 比如8KB的span用来分配32字节的object, 那么就有256个.
	nelems uintptr // number of object in the span.

	// Cache of the allocBits at freeindex. allocCache is shifted
	// such that the lowest bit corresponds to the bit freeindex.
	// allocCache holds the complement of allocBits, thus allowing
	// ctz (count trailing zero) to use it directly.
	// allocCache may contain bits beyond s.nelems; the caller must ignore
	// these.
	//allocBits从第freeindex个位开始的cache. 也就是说, 0-freeindex个之间的对象是被分配了的, 那么要寻找没被分配的位置的对象, 那么就查看allocCache中为0的那一位.
	allocCache uint64

	//sweep之后, 每个slot的分配情况.
	allocBits  *gcBits
	/gc时用于标记哪些slot是被使用的, sweep时赋值给allocBits, gcBits则重新分配空的.
	gcmarkBits *gcBits

	allocCount  uint16     // number of allocated objects
	spanclass   spanClass  // size class and noscan (uint8)
	incache     bool       // being used by an mcache
	state       mSpanState // mspaninuse etc
}
```
## 为什么使用类bitmap的方式标记mspan的分配情况? 不用freelist?
mspan中基本可理解为, 一个对象slot会使用一个bit来标记是否分配.
bitmap中寻找下一个可用的slot会比较麻烦, golang中为了这一点做了比较多的优化, 包括freeindex, allocCache, ctz64等, 算法还比较复杂.
那为啥不用freelist呢? 可用的直接就是链表的下一个, 而指针直接就用该对象所占的内存存储. 也不会造成多的内存需要.
其实在golang1.6之前mspan里空闲对象就是使用freelist链起来的, 在1.7 release中, 改成了现在的类似bitmap的方式.
![golang 1.6的mspan结构体](http://km.oa.com/files/photos/pictures//20190302//1551502040_35.png)

这个golang的gc有关, go在标记时需要记录某个记录是否被使用了. 这样就可以把未被标记的对象释放, 这就是前面讲的bitmap中的markedBit的作用, 
而freelist则是另外一种表示对象是否被使用的方式, 之前的实现中, 相同的事情要做两遍, 需要把bitmap的形式, 转换为freelist的形式.
而且freelist的内存局部性不好.
这个问题有一个issuse和proposal, 并在1.7中改成了bitmap的形式.

下面我做个简单的对比: mspan分配结构

|   freelist  | 类似bitmap    |
| ------------ | ------------ | ------------ |
| 版本| 1.?~1.6(包含)|1.7(包含)~   |
|  mspan初始被分配时 |需要构造freelist  |无需 ,默认都是0 |
| mspan分配对象时 |freelist指针的下一个位置. 这次分配之后, 为优化性能, 会进行后面数据的prefetch  |  结合freeindex和allocCache, 较为复杂, 会使用ctz指令|
|sweep span时(GC标记之后需要清扫mspan)|把代表堆内存bitmap(GC标记之后, 没有被引用到的对象, 不被标记, 即free)的形式转换为freelist的形式(heapBitsSweepSpan), 费时|gc时标记对象直接在mspan中的gcmarkBits中标记, sweep span时不需要重新构建|

### 相关代码提交
**相关的issues**
runtime: replace free list with direct bitmap allocation
https://github.com/golang/go/issues/12800
**相关的proposal**
https://github.com/golang/proposal/blob/master/design/12800-sweep-free-alloc.md
**相关CI和commit**
runtime: bitmap allocation data structs
https://go-review.googlesource.com/c/go/+/19221/
runtime: add bit and cache ctz64 (count trailing zero)
https://go-review.googlesource.com/c/go/+/20200/
**merge**
https://github.com/golang/go/commit/56b54912628934707977a2a0a3824288c0286830
This commit moves the GC from free list allocation to bit mark allocation

# mcache
mcache是每个P带有的用于分配小对象的cache, 因为这个对象是per-P的, 所以访问是不需要锁的.
## P, mcache与mspan的关系图
每个P有67 * 2个mspan作为cache, 一个tiny分配器, 一个多阶的栈内存分配cache.
![p,mcache,mspan的关系](http://km.oa.com/files/photos/pictures//20190302//1551534037_7.png)
##详解版mspan结构体
去除了一些非主要字段. 以下结构体是分配在非Go GC的内存(go heap)中, 所以如果mcache结构体含有go heap中对象的指针的话, 要注意处理, 其实就相当于这个对象作为gc root了, 但是不是分配在go heap中, 所以它不会被扫描. 要特殊处理, 比如tiny分配器.
```go
type mcache struct {

	//这个mcache中分配了多少字节需要scan的对象.
	local_scan  uintptr // bytes of scannable heap allocated

  //用来分配非常小的对象(小于16字节且对象中无指针的tiny object)的cache
  //这个是从size为16bytes的mspan中分配来的, 对于这个对象的标记在_GCmark阶段单独处理一下.
  //tiny分配器合并无指针小对象分配, 简单移动指针, 类似于pointer bumping, 大大加快的分配速度, 这个会在后面单独介绍.
	tiny             uintptr
	//分配到哪个偏移位置
	tinyoffset       uintptr
  //这个mcache用tiny分配器分配了多少对象.
	local_tinyallocs uintptr

	// 以下对象并不是每次分配都涉及到

 //mcache会缓存67*2个不同size class的mspan(当然没有用到前, 是nil的)用于无锁分配内存. per-P的形式, 减少冲突.
	alloc [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass

//分配栈内存也是一样, 尽量减少多线程冲突, per-P做一下cache. 在linux上_NumStackOrders为4. 表示不同size(2KB,4KB,8KB,16KB )的free stack的链表.
	stackcache [_NumStackOrders]stackfreelist
	....
}
```

#mcentral
##mcentral与mspan的关系图
之前讨论到每个P都有一个mcache, 每个mcache有67 \* 2个mspan作为per-P cache, 分配时先用mcache里的mspan, 如果某个size的mspan用完了, 那么如何获取?
这里就涉及到mcentral了, 每个类型的mspan在全局都有一个mcentral作为后备, 用于获取该size类型的mspan.
为什么要有mcache? 不直接从heap里分配呢?
一个是mcentral作为某个size类型的mspan的缓冲器, 二是分成67\*2 个大大减少了多线程冲突.
![mcentral](http://km.oa.com/files/photos/pictures//20190302//1551534225_62.png)
## 详解mcentral结构体
```go
type mcentral struct {
   //因为mentral是全局的, 有可能同时有多个与P绑定的G分配内存时需要申请某个size类型的mspan, 所以需要加锁.
	lock      mutex
	//这个mcentral作为那个size的span的后备. (spanClass是个int类型)
	spanclass spanClass
	//该mcentral中有可以分配对象的mspan构成的list, 如果有mcache需要该mcentral对应的mspan, 那么可以从这个list中分配
	nonempty  mSpanList // list of spans with a free object, ie a nonempty free list
	//表示这个mcentral中满的mspan构成的list
	empty     mSpanList // list of spans with no free objects (or cached in an mcache)

	// nmalloc is the cumulative count of objects allocated from
	// this mcentral, assuming all spans in mcaches are
	// fully-allocated. Written atomically, read under STW.
	nmalloc uint64
}
```

#mheap
##mheap与其他结构体的关系
mheap全局只有一个, runtime/mheap.go中定义的全局变量. 一些访问是需要加锁的.
```go
var mheap_ mheap
```
mheap主要有
1. 全局的mcentral(67\*2个),  page到span的映射, 每个word对应的bitmap.
2. 负责为go runtime从系统申请内存及维护内存位置
3. 负责维护go runtime目前为go程序分配到的位置
4. 为mcentral提供mspan的分配
5. 为>=32KB页直接提供分配
6. 以类似于伙伴系统的形式维护空闲页
7. 还有一些内存分配的统计计数.


mcentral归还给mheap的mspan构成的不同size的连续page, 1-127(\_MaxMHeapList)个page, 即8KB~(1MB-8KB), 类似于伙伴系统. 
如果需要新分配mspan, 那么从其中找到合适大小的span(allocSpanLocked), 有多余的空间切出来, 归还到对应的list中. 如果找不到多余的span, 那么就扩展堆, 从系统申请.
如果有mspan归还, 通过这个span前后的page数找到对应的span, 如果前后的span是free的, 那么进行合并归还, 并插到相应的list.(mheap.freeSpanLocked)

![mheap结构](http://km.oa.com/files/photos/pictures//20190302//1551537449_71.png)
#mheap结构体详解
```go
type mheap struct {
	lock      mutex
	//类似伙伴系统的空闲page的span列表(1-127个page)
	free      [_MaxMHeapList]mSpanList // free lists of given length up to _MaxMHeapList
	//连续page数大于等于128个page的span的树
	freelarge mTreap                   // free treap of length >= _MaxMHeapList
	//所有分配过的mspan结构体, 这个是在分配mspan这个结构体时记录的, 后面会提到fixalloc.
	busy      [_MaxMHeapList]mSpanList // busy lists of large spans of given length

	busylarge mSpanList                // busy lists of large spans length >= _MaxMHeapList
	sweepgen  uint32                   // sweep generation, see comment in mspan
	sweepdone uint32                   // all spans are swept
	sweepers  uint32                   // number of active sweepone calls


	//所有分配过的mspan结构体, 这个是在分配mspan这个结构体时记录的, 后面会提到fixalloc.
	allspans []*mspan // all spans out there


	//前面有提到spans区域, 就是一个以page的index来对应这个page属于那个span.
	//而span本身就有start和npages, 可以找到span有哪些page.
	//这样就建立span和page的双向索引
	spans []*mspan

	//sweepSpans包含了两个mspan构成的stack: 一个是正在使用的已经清扫的span的stack, 一个是正在用的未清扫的span的stack.
	//每次GC, 这两者都会交换身份. 因为sweepgen在每个GC周期增加2.
	//sweepSpans[sweepgen/2%2] 为已经清扫的span
	// sweepSpans[1-sweepgen/2%2]为未清扫的span
	//gc清扫时, 从未清扫的span的stack取出来, 然后清扫, 放到已清扫的span的stack中.
	sweepSpans [2]gcSweepBuf


	//一些分配统计
	// Malloc stats.
	//历史为large object而分配的字节数, 也就是前面讲到的分配大于32KB的对象
	largealloc  uint64                  // bytes allocated for large
	//进行了多少次large alloc
	nlargealloc uint64                  // number of large object allocations
	//历史为>=32KB对象归还heap的字节数
	largefree   uint64                  // bytes freed for large objects (>maxsmallsize)
	//历史>=32KB的large object归还次数
	nlargefree  uint64                  // number of frees for large objects (>maxsmallsize)
	//各个class的小对象的free次数
	nsmallfree  [_NumSizeClasses]uint64 // number of frees for small objects (<=maxsmallsize)

	// range of addresses we might see in the heap
	//记录前面提到的go内存bitmap区域的地址
	bitmap        uintptr // Points to one byte past the end of the bitmap
	bitmap_mapped uintptr


	//go gc heap区域的开始地址
	arena_start uintptr
	//记录bitmap. spans这些辅助区域映射到的位置
	arena_used  uintptr // Set with setArenaUsed.

	// The heap is grown using a linear allocator that allocates
	// from the block [arena_alloc, arena_end). arena_alloc is
	// often, but *not always* equal to arena_used.
	//arena区域目前分配出去的大小
	arena_alloc uintptr
	//arena区域目前从系统申请的大小
	arena_end   uintptr

	arena_reserved bool


	//不同class的mcentral
	central [numSpanClasses]struct {
		mcentral mcentral
		pad      [sys.CacheLineSize - unsafe.Sizeof(mcentral{})%sys.CacheLineSize]byte
	}
	//用于分配mspan结构体本身的固定分配器
	spanalloc             fixalloc // allocator for span*
	//用于分配mcache结构体本身的固定分配器
	cachealloc            fixalloc // allocator for mcache*
	//分配treap节点的固定分配器, 在组织>=128页的连续page时用到
	treapalloc            fixalloc // allocator for treapNodes* used by large objects

}
```


# 几种特殊分配器
## tiny分配器
前面有介绍过, 对于小于16字节的noscan的对象, 使用mcache上的tiny分配器进行分配. 有以下两个原因.

1. 类似于pointer bumping的方式, 分配对象只需要移动offset就可以, 非常的快
2. 小对象整合在一起分配, 提高内存利用率.

目前一个tiny分配器本身16字节, 从16字节对应的class的span中分配. 当前mcache的tiny分配器用完了, 重新分配一个即可. 

![tiny分配器](http://km.oa.com/files/photos/pictures//20190224//1551010580_77.png)

>注意上图只是一个示意图, 其中有一些错误, obj3对象其实是要对齐4字节的.

## fixalloc
我们一直都在讨论go gc heap中对象的分配, 对于管理go heap的一些数据结构, mspan, mcache, mcentral等是从哪里分配的呢?
这些对象就只有几种类型, 而且每个对象都是固定的. 且不应该在go gc heap中分配. 这就是fixalloc的作用. 

定义在runtime/mfixalloc.go文件中. 

```go
type fixalloc struct {
	size   uintptr //分配的对象大小
	first  func(arg, p unsafe.Pointer) // called first time p is returned
	arg    unsafe.Pointer
	list   *mlink //空闲对象构成的链表
	chunk  uintptr //从persistentalloc分配器获取的16K块
	nchunk uint32 //当前chunk块还剩余多少字节
	inuse  uintptr //这个fixalloc一共分配出去多少字节
	stat   *uint64
	zero   bool //分配出去的对象是否需要清零
}
//在mheap中定义以下几个mfixalloc
type mheap{
	//用于分配mspan结构体本身的固定分配器
	spanalloc             fixalloc // allocator for span*
	//用于分配mcache结构体本身的固定分配器
	cachealloc            fixalloc // allocator for mcache*
	//分配treap节点的固定分配器, 在组织>=128页的连续page时用到
	treapalloc            fixalloc // allocator for treapNodes* used by large objects
}
```
> 注意fixalloc分配出去的对象不在go gc heap中, 内存是persistentalloc从系统mmap获取的, 也不由gc释放.


## stackcache
在linux amd64中go协程的栈从2KB开始, 最大可到1GB, 当然大部分栈只会是2KB, 调用过深时扩容到4KB, 8KB...如果每次go一个func就重新申请栈, 那么这个内存分配压力也会比较大. 仿照go gc heap对象分配的原型, 在per-P的mcache中也有多阶的stack列表的cache.

mcache::mcache中的stackcache管理不同规格class的stack, 相同class的stack被链接到同一个链表中.
stackpool:全局stack cache, 和mcache中的stackcache结构相同.
stackLarge: 全局stack cache, 和mcache中的stackcache结构相同. 不同的是stackLarge中stack内存的规格.
```go
//runtime/mcache.go
type struct mcache{
	...
	//linux amd64中为4阶, 分别为2KB, 4KB, 8KB, 16KB.
	stackcache [_NumStackOrders]stackfreelist
	...
}
//runtime/mcache.go
type stackfreelist struct {
	list gclinkptr // linked list of free stacks
	size uintptr   // total size of stacks in list
}

//协程需要栈时, 对于2-16KB的栈, 优先从per-P的stackcache中获取, 没有的话, 再从全局的stackpool申请内存, fill这个cache.
//runtime/stack.go的stackalloc方法.
			x = c.stackcache[order].list
			if x.ptr() == nil {
				stackcacherefill(c, order)
				x = c.stackcache[order].list
			}
			c.stackcache[order].list = x.ptr().next
			c.stackcache[order].size -= uintptr(n)
		}
		v = unsafe.Pointer(x)
//对于大于16KB的stack的申请, 还有一个全局的stackLarge的分配器, 也类似于伙伴系统.
```
stackcache也不过多展开. 相关代码在runtime/stack.go和runtime/mcache.go中.
![全局stackpool结构](http://km.oa.com/files/photos/pictures//20190303//1551620984_70.png)

# 源码分布
本文讨论的大部分代码在以下文件中, 都在runtime包下.
|文件   |内容   |
|---|---|---|
|malloc.go |内存结构的初始化, 内存分配主流程  |
| mheap.go  |mheap定义, mspan定义;从mheap中分配mspan, arean, spans区域的管理  |
|mcentral.go  | mcentral定义;从mcentral中获取span, 或者归还span的方法 |
|mfixalloc.go|fixalloc的定义及相关方法|
|stack.go|栈内存管理|

# 总结
本文从new开始, 先简单分析go对象分配的流程. 然后深入golang runtime包下和内存分配的代码, 详细了解golang runtime的内存结构, mspan, mcentral,mcache, mheap等一些重要的结构体. 最后介绍用于分配小对象的tiny分配器, 分配管理内存的一些结构体的fixalloc, 管理协程栈的stackcache分配器等几个特殊的分配器.

在整个过程中我们知道了, new,make等关键字最终都会转到mallocgc这个方法. golang内存分配以TCMalloc的结构进行设计, 不同大小的对象以不同class的span进行分配,减少内存碎片, 采用per-P的cache减少锁冲突. 同时还有zero分配采用同一个全局对象, tiny分配器加快小对象分配的优化. 对于stack的管理, 也采用类似的结构, 多级cache, 加快分配, 减少gc压力.

# 参考文档
https://blog.learngoprogramming.com/a-visual-guide-to-golang-memory-allocator-from-ground-up-e132258453ed
https://about.sourcegraph.com/go/gophercon-2018-allocator-wrestling
https://go-review.googlesource.com/c/go/+/19221/

Golang源码探索(三) GC的实现原理 https://www.cnblogs.com/zkweb/p/7880099.html

也谈goroutine调度器 https://tonybai.com/2017/06/23/an-intro-about-goroutine-scheduler/

图解 TCMalloc https://zhuanlan.zhihu.com/p/29216091