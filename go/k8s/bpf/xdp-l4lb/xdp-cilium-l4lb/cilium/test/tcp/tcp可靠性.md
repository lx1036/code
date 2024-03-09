
# TCP 可靠性


## 重传


## 流量控制
流量控制：TCP 发送方根据接收方实际接收能力，控制发送的数据量大小。

### 滑动窗口
TCP 原始设计：每次发送一个数据(TCP Segment)，等待接收方应答 ACK 报文，然后才下一个数据发送。这种一问一答模式，效率太低。所以，TCP 需要加滑动窗口设计。
滑动窗口：client 无需等待 server 应答 ACK 报文，在滑动窗口 window scale 内，可以继续发送数据报文。

窗口的实现实际上是操作系统开辟的一个缓存空间，发送方主机在等到确认应答返回之前，必须在缓冲区中保留已发送的数据。
如果按期收到确认应答，此时数据就可以从缓存区清除。




## 拥塞控制
拥塞窗口 cwnd：是发送方维护的一个的状态变量，它会根据网络的拥塞程度动态变化的。一般拥塞控制 congestion control 算法是 cubic。







# TCP/UDP 收包流程
https://zhuanlan.zhihu.com/p/430961897
https://mp.weixin.qq.com/s/pJ2_w3QBTRZG4wK-VI7ZLQ
https://mp.weixin.qq.com/s/6c0ZZ3ZZZ_ocIqH2iey1lw

## 收包流程
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

* 6.数据包进入用户态查找对应的 socket，这里的 socket lookup 逻辑：先查找 established socket，然后查找 listening socket，最后 ANY_ADDR listening socket，
同时第一步和第二步有 ebpf sk_lookup hook，最后找到对应的 socket；

* 7.用户态程序调用 socket 相关 api，如 recvmsg() 或者 recvfrom() 函数获取数据报文；


