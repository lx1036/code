
# TCP 可靠性

## TCP 基本概念
* MTU：一个网络包的最大长度，以太网中一般为 1500 字节。MTU = TCP 头部 + TCP 数据，也就是 IP payload 大小。
这看起来井然有序，但这存在隐患的，那么当如果一个 IP 分片丢失，整个 IP 报文的所有分片都得重传。因为 IP 层本身没有超时重传机制，它由传输层的 TCP 来负责超时和重传。
因此，可以得知由 IP 层进行分片传输，是非常没有效率的。

* MSS：除去 IP 和 TCP 头部之后，一个网络包所能容纳的 TCP 数据的最大长度。也就是 TCP payload 大小。
所以，为了达到最佳的传输效能 TCP 协议在建立连接的时候通常要协商双方的 MSS 值，当 TCP 层发现数据超过 MSS 时，
则就先会进行分片，当然由它形成的 IP 包的长度也就不会大于 MTU ，自然也就不用 IP 分片了。经过 TCP 层分片后，如果一个 TCP 分片丢失后，
进行重发时也是以 MSS 为单位，而不用重传所有的分片，大大增加了重传的效率。

* TIME_WAIT 状态: 主动关闭连接的，才有 TIME_WAIT 状态。
  * 客户端收到服务端的 FIN 报文后，回一个 ACK 应答报文，之后进入 TIME_WAIT 状态。
  * 客户端在经过 2MSL 一段时间后，自动进入 CLOSE 状态，至此客户端也完成连接的关闭。
  * 为什么需要 TIME_WAIT 状态：防止历史连接中的数据，被后面相同四元组的连接错误的接收；




## 流量控制
流量控制：TCP 发送方根据接收方实际接收能力，控制发送的数据量大小。

### 滑动窗口
https://xiaolincoding.com/network/3_tcp/tcp_feature.html

TCP 滑动窗口：有了滑动窗口大小余量，就无需等待 ack 就可以继续发送数据。否则，就每次发送一个数据包，等待对应的 ack 应答，随着报文数据越长，效率就越低。
窗口大小由 TCP header 里的 window 字段确认。

TCP 原始设计：每次发送一个数据(TCP Segment)，等待接收方应答 ACK 报文，然后才下一个数据发送。这种一问一答模式，效率太低。所以，TCP 需要加滑动窗口设计。
滑动窗口：client 无需等待 server 应答 ACK 报文，在滑动窗口 window scale 内，可以继续发送数据报文。

窗口的实现实际上是操作系统开辟的一个缓存空间，发送方主机在等到确认应答返回之前，必须在缓冲区中保留已发送的数据。
如果按期收到确认应答，此时数据就可以从缓存区清除。



## 拥塞控制
拥塞窗口 cwnd：是发送方维护的一个的状态变量，它会根据网络的拥塞程度动态变化的。一般拥塞控制 congestion control 算法是 cubic。
流量控制是避免「发送方」的数据填满「接收方」的缓存，但是并不知道网络的中发生了什么，所以需要拥塞控制来控制 发包速率。

在网络出现拥堵时，如果继续发送大量数据包，可能会导致数据包时延、丢失等，这时 TCP 就会重传数据，但是一重传就会导致网络的负担更重，
于是会导致更大的延迟以及更多的丢包，这个情况就会进入恶性循环被不断地放大。所以，就有了拥塞控制，控制的目的就是避免「发送方」的数据填满整个网络。

拥塞控制解决方案：
* 慢启动: TCP 在刚建立连接完成后，首先是有个慢启动的过程，这个慢启动的意思就是一点一点的提高发送数据包的数量，如果一上来就发大量的数据，这会给网络添堵。
* 拥塞避免算法：当拥塞窗口 cwnd 「超过」慢启动门限 ssthresh 就会进入拥塞避免算法。那么进入拥塞避免算法后，它的规则是：每当收到一个 ACK 时，cwnd 增加 1/cwnd。
* 拥塞发生: 
  * 超时重传
  * 快速重传
* 快速恢复:



# TCP/UDP 收包流程
https://zhuanlan.zhihu.com/p/430961897
https://mp.weixin.qq.com/s/pJ2_w3QBTRZG4wK-VI7ZLQ
https://mp.weixin.qq.com/s/6c0ZZ3ZZZ_ocIqH2iey1lw

## 收包流程
网卡 > ring buffer > skb_buffer > xdp bpf hook > ip_recv() > prerouting netfilter > 
route decision: 
-> ip_forward()
-> ip_local_deliver() > input netfilter > tcp_v4_recv() > lookup established socket > sk_lookup bpf > lookup listening socket > lookup any_addr listening socket

* 1.到达网卡 NIC，通过 DMA(网卡可以不通过CPU访问系统内存) 把 数据帧 在系统内存中，分配环形缓冲区 ring buffer，并网卡验证 MAC 地址；

* 2.触发硬中断，为数据包分配一个 skb_buffer 缓冲区；

* 3.触发软中断，触发网卡驱动程序，收包把数据从 ring buffer 中拷贝到 skb_buffer 缓冲区中，数据送到三层协议栈，见函数 net_rx_action() /root/linux-5.10.142/net/core/dev.c；
* 3.1 数据帧进入三层协议栈前，会经过 xdp ebpf hook；
* 3.2 数据帧进入三层协议栈前，会经过 netif_receive_skb() 函数 /root/linux-5.10.142/net/core/dev.c，它是数据链路层接收数据帧的最后一关；

* 4.对于 IP 协议数据包来说(也可能是arp协议)，调用 ip_rcv() 函数 /root/linux-5.10.142/net/ipv4/ip_input.c，进入三层协议栈。
先 ip hdr 检查和 checksum 检查。然后调用 netfilter NF_INET_PRE_ROUTING hook 中的规则逻辑，是否需要丢弃或者修改数据包。如果 route decision 需要本机处理，
进入 ip_local_deliver()，否则进入 ip_forward() 函数做转发出去；

* 5.对于 TCP 协议数据包，调用 tcp_v4_rcv() 函数，进入四层协议栈。先 tcp hdr 检查和 checksum 检查。然后调用 netfilter
INPUT hook 中的规则逻辑，是否需要丢弃或者修改数据包。对于 UDP 协议数据包，过程类似；

* 6.数据包进入 用户态 查找对应的 socket，这里的 socket lookup 逻辑：先查找 established socket，然后查找 listening socket，最后 ANY_ADDR listening socket，
同时第一步和第二步有 ebpf sk_lookup hook，最后找到对应的 socket；

* 7.用户态程序调用 socket 相关 api，如 recvmsg() 或者 recvfrom() 函数获取数据报文；


