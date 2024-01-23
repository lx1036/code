


# 问题
本 demo 来自于内核 patch 文档 https://www.spinics.net/lists/netdev/msg403579.html, 用来验证 bpf ipip 的示例。

代码在内核里: /root/linux-5.10.142/samples/bpf/tc_l2_redirect.sh

(1)为方便使用，用户态直接使用 bpftool 来 update map, 参考文档 https://manpages.ubuntu.com/manpages/focal/en/man8/bpftool-map.8.html :

```shell
# 安装 bpftool 工具
apt install -y linux-tools-5.4.0-164-generic jq

root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# cat /sys/class/net/tun1/ifindex
45
# 注意这里的 45 值在最前面，而不是 "0 0 0 45", 754974720=int("0x2d000000", 16) [python3]
# 另外 key 和 value 都是四字节大小
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map show pinned /sys/fs/bpf/tc/globals/tun_iface
140: array  flags 0x0
        key 4B  value 4B  max_entries 1  memlock 4096B
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map update pinned /sys/fs/bpf/tc/globals/tun_iface key 0 0 0 0 value 45 0 0 0
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map dump pinned /sys/fs/bpf/tc/globals/tun_iface -j | jq
[
  {
    "key": [
      "0x00",
      "0x00",
      "0x00",
      "0x00"
    ],
    "value": [
      "0x2d",
      "0x00",
      "0x00",
      "0x00"
    ]
  }
]

```

(2)查看 bpf_trace_printk() 函数的日志
bpf_trace_printk() 只能带3个参数。
```shell
tail -n 100 /sys/kernel/debug/tracing/trace
# tracer: nop
#
# entries-in-buffer/entries-written: 5/5   #P:4
#
#                                _-----=> irqs-off
#                               / _----=> need-resched
#                              | / _---=> hardirq/softirq
#                              || / _--=> preempt-depth
#                              ||| /     delay
#           TASK-PID     CPU#  ||||   TIMESTAMP  FUNCTION
#              | |         |   ||||      |         |
            ping-1137009 [001] ..s1 150505.398215: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137062 [001] ..s1 150515.532237: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137518 [002] ..s1 150769.460319: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137844 [003] ..s1 150923.706842: 0: e/ingress redirect daddr4:a0a0166 to ifindex:45
     ksoftirqd/3-30      [003] ..s. 150923.706959: 0: ingress forward to ifindex:45 daddr4:a020101
```

# mac 本地使用 bpf_trace_printk() debug ebpf
mac 本地运行，需要 volume debugfs(验证可用)
```shell
docker volume create --driver local --opt type=debugfs --opt device=debugfs debugfs

docker stop ebpf-for-mac && docker rm ebpf-for-mac

docker run -it --name ebpf-for-mac --privileged -v debugfs:/sys/kernel/debug:ro \
-v /lib/modules:/lib/modules:ro -v /etc/localtime:/etc/localtime:ro --pid=host \
-v /Users/liuxiang/Code/code:/mnt/code \
-v /Users/liuxiang/go/pkg/mod:/root/go/pkg/mod \
lx1036/ebpf-for-mac:2.1 /bin/bash
```


# 过程解释

过程描述：ns1 里 ping vip(10.10.1.102) 是通的

```
root@xxx:~/ipip# ip netns exec ns1 ping -c3 10.10.1.102
PING 10.10.1.102 (10.10.1.102) 56(84) bytes of data.
64 bytes from 10.10.1.102: icmp_seq=1 ttl=63 time=0.064 ms
64 bytes from 10.10.1.102: icmp_seq=2 ttl=63 time=0.069 ms
64 bytes from 10.10.1.102: icmp_seq=3 ttl=63 time=0.067 ms

--- 10.10.1.102 ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2041ms
rtt min/avg/max/mdev = 0.064/0.066/0.069/0.002 ms
```

解释包的流程：
路由决策 -> vens1(ns1) ->[veth pair]-> ve1(tc ingress redirect, host) -> bpf_redirect -> vens2(ns2) 已经做了 ipip 封装  -> 
路由决策 -> ve2(host) -> ve2(host, tc ingress forward) -> tun1(host) ->
路由决策 -> vens2(ns2) -> vens2(tc ingress) -> tun2(tcpdump 后已经 ipip 解包)

ping (10.1.1.101 > 10.10.1.102) -> veth1(l2_to_iptun_ingress_redirect) -> tun1 egress -> (10.1.1.101 > 10.10.1.102) ->
veth2 10.1.2.1 > 10.1.2.101(10.1.1.101 > 10.10.1.102), 外层 iphdr dstIP 是 l2_to_iptun_ingress_redirect 添加的, 10.1.2.1 srcIP 是路由添加的 -> 
eth0(ns2) 10.1.1.1 > 10.1.2.101(10.1.1.101 > 10.10.1.102) ->
tun2 命中 local/remote 解包 (10.1.1.101 > 10.10.1.102) -> lo (10.10.1.102 > 10.1.1.101) -> 命中路由 -> tun2 10.10.1.102 > 10.1.1.101 ->
tun2 命中 local/remote 封包 10.1.2.101 > 10.1.2.1(10.10.1.102 > 10.1.1.101) -> veth2 10.1.2.101 > 10.1.2.1(10.10.1.102 > 10.1.1.101) ->
veth2(l2_to_iptun_ingress_forward) -> tun1 ingress -> tun1 10.2.1.102 > 10.2.1.1(10.10.1.102 > 10.1.1.101) -> tun1 ipip external 解包外层 iphdr ->
10.10.1.102 > 10.1.1.101 -> 路由 -> veth1 10.10.1.102 > 10.1.1.101 -> eth0(ns1) 10.10.1.102 > 10.1.1.101

> tcpdump 在 xdp 之后，在 tc ingress 之前抓包，所以 tcpdump -i veth1 会在 tc ingress rule 里 redirect 之前能抓到包。 

```
# tcpdump -i veth1 -nneevv -A icmp

# 包从 veth1 -> tun1 egress，会走 xdp->tcpdump->tc_ingress->netfilter->tc_egress，绕过 tc_ingress->netfilter，
# 所以能抓到包，且去包和 veth1 一样都是 icmp 包，没有被 tun1 封包
# tcpdump -i tun1 -nneevv -A
# tcpdump -i veth2 -nneevv -A
# ip netns exec ns2 tcpdump -i eth0 -nneevv -A
# ip netns exec ns2 tcpdump -i tun2 -nneevv -A

# ip netns exec ns2 tcpdump -i eth0 -nneevv -A
16:45:12.749308 16:01:bc:fe:59:75 > d2:b4:19:fa:f2:ae, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 29564, offset 0, flags [none], proto IPIP (4), length 104)
    10.2.1.1 > 10.2.1.102: (tos 0x0, ttl 64, id 1499, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.101 > 10.10.1.102: ICMP echo request, id 16026, seq 1, length 64

16:45:12.749437 d2:b4:19:fa:f2:ae > 16:01:bc:fe:59:75, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 7393, offset 0, flags [DF], proto IPIP (4), length 104)
    10.2.1.102 > 10.2.1.1: (tos 0x0, ttl 64, id 49789, offset 0, flags [none], proto ICMP (1), length 84)
    10.10.1.102 > 10.1.1.101: ICMP echo reply, id 16026, seq 1, length 64


# ip netns exec ns2 tcpdump -i tun2 -nneevv -A
17:30:14.818789 ip: (tos 0x0, ttl 64, id 5112, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.101 > 10.10.1.102: ICMP echo request, id 16865, seq 1, length 64

17:30:14.818814 ip: (tos 0x0, ttl 64, id 38308, offset 0, flags [none], proto ICMP (1), length 84)
    10.10.1.102 > 10.1.1.101: ICMP echo reply, id 16865, seq 1, length 64
```


# 问题

## 2024-01-17
这次遇到的问题，是没有 ping 通，原因是 /proc/sys/net/ipv4/conf/tun1/rp_filter=2，应该为 0:
```shell
# rp_filter 是Linux内核中的一个参数，用于控制网络包的接收策略。它主要用于防止IP欺骗攻击，
# 通过检查接收到的数据包是否来自正确的源地址来决定是否接受该数据包
# 0：关闭反向路径过滤功能。
# 1：只检查本地子网内的数据包。如果数据包的目的地址不在本地子网内，则直接丢弃该数据包。
# 2：检查所有的数据包。如果数据包的目的地址不在本地子网内，并且没有有效的路由可以到达该地址，则直接丢弃该数据包。
sysctl -q -w net.ipv4.conf.tun1.rp_filter=0 # 2 不可以，但是还是没法解释
```

## tcpdump 和 ebpf hook 流程
xdp->tcpdump->tc_ingress->netfilter->tc_egress: https://arthurchiao.art/blog/firewalling-with-bpf-xdp/#51-tcpdump-hooking-point


# 参考文献

https://www.spinics.net/lists/netdev/msg403578.html
https://www.spinics.net/lists/netdev/msg403580.html
https://www.spinics.net/lists/netdev/msg403579.html

