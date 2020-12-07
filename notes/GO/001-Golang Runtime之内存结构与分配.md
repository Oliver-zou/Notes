[TOC]
# 简介
本文先介绍PMG模型作为基础. 以new一个对象为引子, 整体介绍Golang Runtime中内存结构和内存分配, 想要深入了解的话, 本文在后面给出了很多不错的深入解析文章. 有问题欢迎企业微信yifhao讨论.
## 章节安排
本文整体分为三大部分.
第一部分介绍一点golang PMG调度模型的背景知识, 以及与内存分配的关系.
第二部分以new一个对象为引子, 简单介绍golang中的内存分配流程和策略.
第三部分深入解析go的内存结构和内存分配, 介绍mspan, mcache, mcentral, mheap等数据结构, 小于16Byte, 16Byte-32KB, >=32KB内存分配流程, 以及tiny分配器, fixalloc分配器, stack等几个特殊分配器.
.
# PMG调度模型简介
## PMG简介
这里简单介绍一下Go的调度模型, 这个调度模型与Go的内存分配结构有很大的联系.
G是协程,  非常轻量级,  初始化栈大小仅2KB(不同系统不同).
P代表一个逻辑Processor, 是资源的拥有者, 包含了per-P的内存分配资源, 本地运行队列等, P的个数限制最大可同时运行的G的个数, 初始化时设置为核的个数或者自定义的GOMAXPROCS变量. per-P数据结构的访问是无锁的.
M则对应了一个系统线程, 如果要运行Go代码(而不是系统调用或者cgo), M需要与一个P结合, 通过调度寻找可运行的G, 真正执行G中的func.
![PMG调度图](http://km.oa.com/files/photos/pictures//20181127//1543322454_59.png)
## 为什么要有PMG
**为什么要有PMG模型? MG不就够了吗? PMG和内存有什么联系?**

在Go 1.0版本中, 实现的是MG模型, 存在以下问题:

1. 调度锁问题. 单一的全局调度锁(Sched.Lock)和集中的状态, 导致伸缩性下降. 所有和goroutine相关比如创建, 完成, 重新调度等操作都要争抢这把锁..
2. G传递问题. 在工作线程M之间需要经常传递runnable的G, 这个会加大调度延迟, 并带来额外的性能损耗.
3. **Per-M的内存问题. 基于TCMalloc结构的内存结构, 每个M都需要memory cache和其他类型的cache(stack alloc), 然而实际上只有M在运行Go代码时才需要这些Per-M Cache, 阻塞在系统调用的M并不需要mcache. 正在运行Go代码的M与进行系统调用的M的比例可能高达1:100, 这造成了极大的内存消耗, 和很差的数据局部性**
4. 线程阻塞与从阻塞恢复. 由于系统调用而形成的剧烈的worker thread阻塞和解除阻塞，导致很大的性能损耗.

Dmitry Vyukov提出和实现了PMG模型(Go 1.2)
Scalable Go Scheduler Design Doc
https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw/edit

**通过P解耦G和M, 将之前与M关联的mcache移至P中. 极大的减少了内存消耗.**
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

等等

# Last
ppt在我的git code上, 这个仓库放的是我在腾讯做的一些分享, 有计划把Golang Runtime的调度, 内存, GC, 并发工具, 重要数据结构实现都整理分享一下,目前已经做了调度, 内存, GC了. 欢迎star. http://git.code.oa.com/yifhao/myshare
同时本文的markdown放在了深入golang运行时的仓库中 https://git.code.oa.com/yifhao/dive-into-golang-runtime, 大家一起来补充和完善. 后续golang GC, 调度, 并发, 数据结构的一些分析文章也会包含在里面.

> 本文可保留署名随意引用和转载, 文章中引用的图片归原作者所有.