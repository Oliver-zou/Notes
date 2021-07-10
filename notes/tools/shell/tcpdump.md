##  应用场景

在日常工作中遇到的很多网络问题都可以通过 tcpdump 优雅的解决：

*1.* 相信大多数同学都遇到过 SSH 连接服务器缓慢，通过 tcpdump 抓包，可以快速定位到具体原因，一般都是因为 DNS 解析速度太慢。

*2.* 当我们工程师与用户面对网络问题争执不下时，通过 tcpdump 抓包，可以快速定位故障原因，轻松甩锅，毫无压力。

*3.* 当我们新开发的网络程序，没有按照预期工作时，通过 tcpdump 收集相关数据包，从包层面分析具体原因，让问题迎刃而解。

*4.* 当我们的网络程序性能比较低时，通过 tcpdump 分析数据流特征，结合相关协议来进行网络参数优化，提高系统网络性能。

*5.* 当我们学习网络协议时，通过 tcpdump 抓包，分析协议格式，帮助我们更直观、有效、快速的学习网络协议。

上述只是简单罗列几种常见的应用场景，而 tcpdump 在网络诊断、网络优化、协议学习方面，确实是一款非常强大的网络工具，只要存在网络问题的地方，总能看到它的身影。

## 工作原理

tcpdump 是 Linux 系统中非常有用的网络工具，运行在用户态，本质上是通过调用 `libpcap` 库的各种 `api` 来实现数据包的抓取功能。

<div align="center"> <img src="../../../pics/20210528152718.jpg" width="300px"/> </div><br>

通过上图，我们可以很直观的看到，数据包到达网卡后，经过数据包过滤器（BPF）筛选后，拷贝至用户态的 tcpdump 程序，以供 tcpdump 工具进行后续的处理工作，输出或保存到 pcap 文件。

数据包过滤器（BPF）主要作用，就是根据用户输入的过滤规则，只将用户关心的数据包拷贝至 tcpdump，这样能够减少不必要的数据包拷贝，降低抓包带来的性能损耗。

**思考**：这里分享一个真实的面试题

> 面试官：如果某些数据包被 iptables 封禁，是否可以通过 tcpdump 抓到包？

通过上图，我们可以很轻易的回答此问题。

因为 Linux 系统中 `netfilter` 是工作在协议栈阶段的，tcpdump 的过滤器（BPF）工作位置在协议栈之前，所以当然是可以抓到包了！

## 实战：基础用法

我们先通过几个简单的示例来介绍 tcpdump 基本用法。

*1.* 不加任何参数，默认情况下将抓取第一个非 lo 网卡上所有的数据包

```
$ tcpdump 
```

*2.* 抓取 eth0 网卡上的所有数据包

```
$ tcpdump -i eth0
```

*3.* 抓包时指定 `-n` 选项，不解析主机和端口名。这个参数很关键，会影响抓包的性能，一般抓包时都需要指定该选项。

```
$ tcpdump -n -i eth0
```

*4.* 抓取指定主机  `192.168.1.100` 的所有数据包

```
$ tcpdump -ni eth0 host 192.168.1.100
```

*5.* 抓取指定主机 `10.1.1.2` 发送的数据包

```
$ tcpdump -ni eth0 src host 10.1.1.2
```

*6.* 抓取发送给 `10.1.1.2` 的所有数据包

```
$ tcpdump -ni eth0 dst host 10.1.1.2
```

*7.* 抓取 eth0 网卡上发往指定主机的数据包，抓到 10 个包就停止，这个参数也比较常用

```
$ tcpdump -ni eth0 -c 10 dst host 192.168.1.200
```

*8.* 抓取 eth0 网卡上所有 SSH 请求数据包，SSH 默认端口是 22

```
$ tcpdump -ni eth0 dst port 22
```

*9.* 抓取 eth0 网卡上 5 个 ping 数据包

```
$ tcpdump -ni eth0 -c 5 icmp
```

*10.* 抓取 eth0 网卡上所有的 arp 数据包

```
$ tcpdump -ni eth0 arp
```

*11.* 使用十六进制输出，当你想检查数据包内容是否有问题时，十六进制输出会很有帮助。

```
$ tcpdump -ni eth0 -c 1 arp -X
listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes
12:13:31.602995 ARP, Request who-has 172.17.92.133 tell 172.17.95.253, length 28
    0x0000:  0001 0800 0604 0001 eeff ffff ffff ac11  ................
    0x0010:  5ffd 0000 0000 0000 ac11 5c85            _.........\.
```

*12.* 只抓取 eth0 网卡上 IPv6 的流量

```
$ tcpdump -ni eth0 ip6
```

*13.* 抓取指定端口范围的流量

```
$ tcpdump -ni eth0 portrange 80-9000
```

*14.* 抓取指定网段的流量

```
$ tcpdump -ni eth0 net 192.168.1.0/24
```

## 实战：高级进阶

tcpdump 强大的功能和灵活的策略，主要体现在过滤器（BPF）强大的表达式组合能力。

本节主要分享一些常见的所谓高级用法，希望读者能够举一反三，根据自己实际需求，来灵活使用它。

*1.* 抓取指定客户端访问 ssh 的数据包

```
$ tcpdump -ni eth0 src 192.168.1.100 and dst port 22
```

*2.* 抓取从某个网段来，到某个网段去的流量

```
$ tcpdump -ni eth0 src net 192.168.1.0/16 and dst net 10.0.0.0/8 or 172.16.0.0/16
```

*3.* 抓取来自某个主机，发往非 ssh 端口的流量

```
$ tcpdump -ni eth0 src 10.0.2.4 and not dst port 22
```

*4.* 当构建复杂查询的时候，你可能需要使用引号，单引号告诉 tcpdump 忽略特定的特殊字符，这里的 `()` 就是特殊符号，如果不用引号的话，就需要使用转义字符

```
$ tcpdump -ni eth0 'src 10.0.2.4 and (dst port 3389 or 22)'
```

*5.* 基于包大小进行筛选，如果你正在查看特定的包大小，可以使用这个参数

小于等于 64 字节：

```
$ tcpdump -ni less 64
```

大于等于 64 字节：

```
$ tcpdump -ni eth0 greater 64
```

等于 64 字节：

```
$ tcpdump -ni eth0 length == 64
```

*6.* 过滤 TCP 特殊标记的数据包

抓取某主机发送的 `RST` 数据包：

```
$ tcpdump -ni eth0 src host 192.168.1.100 and 'tcp[tcpflags] & (tcp-rst) != 0'
```

抓取某主机发送的 `SYN` 数据包：

```
$ tcpdump -ni eth0 src host 192.168.1.100 and 'tcp[tcpflags] & (tcp-syn) != 0'
```

抓取某主机发送的 `FIN` 数据包：

```
$ tcpdump -ni eth0 src host 192.168.1.100 and 'tcp[tcpflags] & (tcp-fin) != 0'
```

抓取 TCP 连接中的 `SYN` 或 `FIN` 包

```
$ tcpdump 'tcp[tcpflags] & (tcp-syn|tcp-fin) != 0'
```

*7.* 抓取所有非 ping 类型的 `ICMP` 包

```
$ tcpdump 'icmp[icmptype] != icmp-echo and icmp[icmptype] != icmp-echoreply'
```

*8.* 抓取端口是 80，网络层协议为 IPv4， 并且含有数据，而不是 SYN、FIN 以及 ACK 等不含数据的数据包

```
$ tcpdump 'tcp port 80 and (((ip[2:2] - ((ip[0]&0xf)<<2)) - ((tcp[12]&0xf0)>>2)) != 0)'
```

解释一下这个复杂的表达式，具体含义就是，整个 IP 数据包长度减去 IP 头长度，再减去 TCP 头的长度，结果不为 0，就表示数据包有 `data`，如果还不是很理解，需要自行补一下 `tcp/ip` 协议

*9.* 抓取 HTTP 报文，`0x4754` 是 `GET` 前两字符的值，`0x4854` 是 `HTTP` 前两个字符的值

```
$ tcpdump  -ni eth0 'tcp[20:2]=0x4745 or tcp[20:2]=0x4854'
```

## 常用选项

通过上述的实战案例，相信大家已经掌握的 `tcpdump` 基本用法，在这里来详细总结一下常用的选项参数。

**（一）基础选项**

- `-i`：指定接口
- `-D`：列出可用于抓包的接口
- `-s`：指定数据包抓取的长度
- `-c`：指定要抓取的数据包的数量
- `-w`：将抓包数据保存在文件中
- `-r`：从文件中读取数据
- `-C`：指定文件大小，与 `-w` 配合使用
- `-F`：从文件中读取抓包的表达式
- `-n`：不解析主机和端口号，这个参数很重要，一般都需要加上
- `-P`：指定要抓取的包是流入还是流出的包，可以指定的值 `in`、`out`、`inout`

**（二）输出选项**

- `-e`：输出信息中包含数据链路层头部信息
- `-t`：显示时间戳，`tttt` 显示更详细的时间
- `-X`：显示十六进制格式
- `-v`：显示详细的报文信息，尝试 `-vvv`，`v` 越多显示越详细

## 过滤表达式

tcpdump 强大的功能和灵活的策略，主要体现在过滤器（BPF）强大的表达式组合能力。

**（一）操作对象**

表达式中可以操作的对象有如下几种：

- `type`，表示对象的类型，比如：`host`、`net`、`port`、`portrange`，如果不指定 type 的话，默认是 host
- `dir`：表示传输的方向，可取的方式为：`src`、`dst`。
- `proto`：表示协议，可选的协议有：`ether`、`ip`、`ip6`、`arp`、`icmp`、`tcp`、`udp`。

**（二）条件组合**

表达对象之间还可以通过关键字 `and`、`or`、`not` 进行连接，组成功能更强大的表达式。

- `or`：表示或操作
- `and`：表示与操作
- `not`：表示非操作

建议看到这里后，再回头去看实战篇章的示例，相信必定会有更深的理解。如果是这样，那就达到了我预期的效果了！

## 经验

到这里就不再加新知识点了，分享一些工作中总结的经验：

*1.* 我们要知道 `tcpdump` 不是万能药，并不能解决所有的网络问题。

*2.* 在高流量场景下，抓包可能会影响系统性能，如果是在生产环境，请谨慎使用！

*3.* 在高流量场景下，`tcpdump` 并不适合做流量统计，如果需要，可以使用交换机镜像的方式去分析统计。

*4.* 在 Linux 上使用 `tcpdump` 抓包，结合 `wireshark` 工具进行数据分析，能事半功倍。

*5.* 抓包时，尽可能不要使用 `any` 接口来抓包。

*6.* 抓包时，尽可能指定详细的数据包过滤表达式，减少无用数据包的拷贝。

*7.* 抓包时，尽量指定 `-n` 选项，减少解析主机和端口带来的性能开销。
