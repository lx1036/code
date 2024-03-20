


# conntrack
linux kernel 内核里 conntrack 做的事情：
* 从数据包中提取 tuple 元组信息 struct nf_conntrack_tuple{}，来辨别 flow 数据流和 connection 链接。注意：CT 中，一个 tuple 表示的一个 flow，意思就是一个 connection；
* 为所有 connection 维护一个 conntrack table，记录 connection 创建时间、send/recv packets、send/recv bytes； 
* GC 回收过期的 connection；
* 为上层服务如 NAT 提供支持；

* struct nf_conntrack_tuple{}: /root/linux-5.10.142/include/net/netfilter/nf_conntrack_tuple.h
* struct nf_conn{}：定义一个 flow
* 入口函数：nf_conntrack_in(struct sk_buff *skb, const struct nf_hook_state *state){}
  * 会存储在一个 hash 表里，key = hash_conntrack_raw(), value = nf_conntrack_tuple_hash()

netfilter 会在四个 hook 点调用 nf_conntrack_in() 从而存储 conntrack tuple 信息：
* PRE_ROUTING 和 LOCAL_OUT
  * 记录一条 connection, 存储在 unconfirmed list
  * 新链接第一个包到达的第一个 hook 点地方：外部包主动到达本机，PRE_ROUTING 是第一个 hook 点；本机包主动到外部，OUTPUT 是第一个 hook 点
* POST_ROUTING 和 LOCAL_IN
  * 调用 nf_conntrack_confirm() 将 unconfirmed list 里的 connection(nf_conntrack_in() 创建的)， 确认为 confirmed list，说明之前 filter 过程没有丢弃这个包；
    * 之所以把创建一个新 entry 的过程分为创建（new）和确认（confirm）两个阶段 ，是因为包在经过 nf_conntrack_in() 之后，到达 nf_conntrack_confirm() 之前 ，可能会被内核丢弃。
    * 这样会导致系统残留大量的半连接状态记录，在性能和安全性上都 是很大问题。分为两步之后，可以加快半连接状态 conntrack entry 的 GC。
  * 离开 netfilter 之后的最后的 hook 点：外部包主动到达本机，LOCAL_IN 是被送到本机进程之前的最后一个 hook 点；本机包主动到外部，POST_ROUTING 是离开本机最后一个 hook 点；
* 总结：外部主动到达本机：PRE_ROUTING > LOCAL_IN ；本机主动到达外部：LOCAL_OUT > POST_ROUTING

## 查看/加载/卸载 nf_conntrack 模块
```shell
# 查看
lsmod | grep nf_conntrack
modinfo nf_conntrack
# 加载
modprobe nf_conntrack
# 加载时还可以指定额外的配置参数，例如：
modprobe nf_conntrack nf_conntrack_helper=1 expect_hashsize=131072
# 卸载
rmmod nf_conntrack
# 查看内核 conntrack 相关参数配置
sysctl -a | grep nf_conntrack
```

## conntrack CLI
```shell
# 内核 netfilter 实现的 conntrack 查看
apt install -y conntrack
# 本地电脑 connect ecs:80
nc 10.20.30.40 80 -v
conntrack -L conntrack
tcp      6 431992 ESTABLISHED src=10.20.30.40 dst=172.16.3.151 sport=65527 dport=80 src=172.16.3.151 dst=10.20.30.40 sport=80 dport=65527 [ASSURED] mark=0 use=2
conntrack v1.4.6 (conntrack-tools): 38 flow entries have been shown.

# cilium ebpf 实现的 conntrack 查看，也可看出 conntrack 就是要存储 创建过期时间、tx/rx packets/bytes、
cilium bpf ct list global
TCP IN 10.244.1.120:60366 -> 10.244.1.170:4240 expires=17335515 RxPackets=50148 RxBytes=4601087 RxFlagsSeen=0x1a LastRxReport=17314419 TxPackets=37611 TxBytes=3422609 TxFlagsSeen=0x1a LastTxReport=17314419 Flags=0x0010 [ SeenNonSyn ] RevNAT=0 SourceSecurityID=1 IfIndex=0 
TCP OUT 192.168.49.3:57686 -> 192.168.49.2:8443 expires=17335517 RxPackets=792140 RxBytes=431973430 RxFlagsSeen=0x18 LastRxReport=17314422 TxPackets=419094 TxBytes=53573408 TxFlagsSeen=0x1a LastTxReport=17314422 Flags=0x0010 [ SeenNonSyn ] RevNAT=0 SourceSecurityID=0 IfIndex=0 
TCP OUT 10.244.1.120:36516 -> 10.244.0.83:4240 expires=17335515 RxPackets=62686 RxBytes=5077642 RxFlagsSeen=0x1a LastRxReport=17314422 TxPackets=37612 TxBytes=3121838 TxFlagsSeen=0x1a LastTxReport=17314422 Flags=0x0010 [ SeenNonSyn ] RevNAT=0 SourceSecurityID=0 IfIndex=0

cilium bpf nat list
UDP IN 192.168.49.2:8472 -> 192.168.49.3:60881 XLATE_DST 192.168.49.3:60881 Created=550658sec HostLocal=1
TCP IN 192.168.49.2:8443 -> 192.168.49.3:57710 XLATE_DST 192.168.49.3:57710 Created=550658sec HostLocal=1
ICMP IN 10.244.0.83:0 -> 10.244.1.120:38778 XLATE_DST 10.244.1.120:38778 Created=550658sec HostLocal=1
```






# 参考文献

**[连接跟踪（conntrack）：原理、应用及 Linux 内核实现](https://arthurchiao.art/blog/conntrack-design-and-implementation-zh)**



