


# 同节点 pod 使用 veth-pair 打通
https://zhuanlan.zhihu.com/p/571966611
不使用 bridge 仅使用 veth-pair 来打通同节点 pod 通信

```shell
# 不使用 bridge 仅使用 veth-pair 来打通同节点 pod 通信
ip netns add ns1
ip netns add ns2
ip link add veth1 type veth peer name eth0 netns ns1
ip link add veth2 type veth peer name eth0 netns ns2
ip link set veth1 up
ip link set veth2 up
ip link set dev veth1 address ee:ee:ee:ee:ee:ee
ip link set dev veth2 address ee:ee:ee:ee:ee:ee

ip netns exec ns1 ip link set eth0 up
ip netns exec ns2 ip link set eth0 up
ip netns exec ns1 ip link set lo up
ip netns exec ns2 ip link set lo up
ip netns exec ns1 ip addr add 100.0.1.1/32 dev eth0
ip netns exec ns2 ip addr add 100.0.1.2/32 dev eth0

ip netns exec ns1 ip route add 169.254.1.1/32 dev eth0
ip netns exec ns1 ip route add default via 169.254.1.1 dev eth0
ip netns exec ns2 ip route add 169.254.1.1/32 dev eth0
ip netns exec ns2 ip route add default via 169.254.1.1 dev eth0
ip netns exec ns1 ip neigh replace 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee
#ip netns exec ns1 ip neigh add 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee
ip netns exec ns2 ip neigh replace 169.254.1.1 dev eth0 lladdr ee:ee:ee:ee:ee:ee

ip route add 100.0.1.1/32 dev veth1
ip route add 100.0.1.2/32 dev veth2

# 必须同时打开 forwarding 和 proxy_arp，开启 veth1/veth2 网卡的 proxy_arp 应答
echo 1 > /proc/sys/net/ipv4/conf/veth1/proxy_arp
echo 1 > /proc/sys/net/ipv4/conf/veth1/forwarding
# 必须同时打开 forwarding 和 proxy_arp
echo 1 > /proc/sys/net/ipv4/conf/veth2/proxy_arp
echo 1 > /proc/sys/net/ipv4/conf/veth2/forwarding

ip netns exec ns1 ping -c 3 100.0.1.2
ip netns exec ns2 ping -c 3 100.0.1.1
```


# 跨节点 pod 使用 ipip 打通

## 抓包验证 ipip(验证通过)
参考自文章: https://blog.csdn.net/qq_22918243/article/details/130886641
自己实现一个 ipip cni: https://zhuanlan.zhihu.com/p/571966611

https://developers.redhat.com/blog/2019/05/17/an-introduction-to-linux-virtual-interfaces-tunnels#ipip
ethhdr -> outer iphdr(ipip proto) -> inner iphdr -> payload : https://datatracker.ietf.org/doc/html/rfc2003

ipip linux 源码: /root/linux-5.10.142/net/ipv4/ipip.c
ipip wiki: https://en.wikipedia.org/wiki/IP_in_IP
ipip shell 命令验证: https://blog.csdn.net/qq_22918243/article/details/130886641


## ipip demo 验证
注意：这里设置也有缺点，`remote 172.16.111.102 local 172.16.111.103` 也是写死的。

```shell
# nodeIP
ecs1 172.16.111.103
ecs2 172.16.111.102

# ipip 里 remote 和 local 用来匹配包，如果不想具体指定，可以设置为 remote any local any

# ecs1
ip tunnel add ipip.1 mode ipip remote 172.16.111.102 local 172.16.111.103
ip link set ipip.1 up
ip addr add 10.10.100.10 peer 10.10.200.10 dev ipip.1
# 会自动创建一条路由
10.10.200.10 dev tun1 proto kernel scope link src 10.10.100.10
# ip addr add INTERNAL_IPV4_ADDR/24 dev ipip.1
# 然后手动创建路由
# ip route add REMOTE_INTERNAL_SUBNET/24 dev ipip.1

# ecs2
ip tunnel add ipip.1 mode ipip remote 172.16.111.103 local 172.16.111.102
ip link set ipip.1 up
ip addr add 10.10.200.10 peer 10.10.100.10 dev ipip.1
# 会自动创建一条路由
10.10.100.10 dev ipip.1 proto kernel scope link src 10.10.200.10

root@xxx:~# ip -d addr show ipip.1
5: ipip.1@NONE: <POINTOPOINT,NOARP,UP,LOWER_UP> mtu 1480 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ipip 172.16.111.103 peer 172.16.111.102
    inet 10.10.100.10 peer 10.10.200.10/32 scope global ipip.1
       valid_lft forever preferred_lft forever
    inet6 fe80::5efe:ac10:6f67/64 scope link 
       valid_lft forever preferred_lft forever

root@xxx:~# ip -d addr show ipip.1
5: ipip.1@NONE: <POINTOPOINT,NOARP,UP,LOWER_UP> mtu 1480 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ipip 172.16.111.102 peer 172.16.111.103
    inet 10.10.200.10 peer 10.10.100.10/32 scope global ipip.1
       valid_lft forever preferred_lft forever
    inet6 fe80::5efe:ac10:6f66/64 scope link 
       valid_lft forever preferred_lft forever

# 验证

# 在 ecs1
root@xxx:~# ping 10.10.200.10
PING 10.10.200.10 (10.10.200.10) 56(84) bytes of data.
64 bytes from 10.10.200.10: icmp_seq=1 ttl=64 time=0.283 ms
64 bytes from 10.10.200.10: icmp_seq=2 ttl=64 time=0.124 ms
64 bytes from 10.10.200.10: icmp_seq=3 ttl=64 time=0.104 ms
64 bytes from 10.10.200.10: icmp_seq=4 ttl=64 time=0.147 ms
64 bytes from 10.10.200.10: icmp_seq=5 ttl=64 time=0.128 ms

# 在 ecs2
root@xxx:~# ping 10.10.100.10
PING 10.10.100.10 (10.10.100.10) 56(84) bytes of data.
64 bytes from 10.10.100.10: icmp_seq=1 ttl=64 time=0.130 ms
64 bytes from 10.10.100.10: icmp_seq=2 ttl=64 time=0.116 ms
64 bytes from 10.10.100.10: icmp_seq=3 ttl=64 time=0.108 ms
64 bytes from 10.10.100.10: icmp_seq=4 ttl=64 time=0.114 ms

```

## tcpdump 抓包 ipip
必须抓包 eth0 端口，而不是 ipip.1 端口:

```shell
# 指定 ip src 和 ip dst
tcpdump -i eth0 -nneevv -A ip src 172.16.111.103 and ip dst 172.16.111.102
tcpdump -i eth0 -nneevv -A host 172.16.111.103

# ip.proto==4 是 IPIP
tcpdump -i eth0 -nneevv -A proto 4

tcpdump: verbose output suppressed, use -v[v]... for full protocol decode
listening on eth0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
17:42:30.335683 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [S], seq 3890488449, win 64800, options [mss 1440,sackOK,TS val 2882874243 ecr 0,nop,wscale 7], length 0
17:42:30.335737 IP 172.16.111.102 > 172.16.111.103: IP 10.10.200.10.123 > 10.10.100.10.37476: Flags [S.], seq 3141031332, ack 3890488450, win 64260, options [mss 1440,sackOK,TS val 1533882345 ecr 2882874243,nop,wscale 7], length 0
17:42:30.335907 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [.], ack 1, win 507, options [nop,nop,TS val 2882874243 ecr 1533882345], length 0
17:42:30.335937 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [P.], seq 1:81, ack 1, win 507, options [nop,nop,TS val 2882874243 ecr 1533882345], length 80
17:42:30.335947 IP 172.16.111.102 > 172.16.111.103: IP 10.10.200.10.123 > 10.10.100.10.37476: Flags [.], ack 81, win 502, options [nop,nop,TS val 1533882345 ecr 2882874243], length 0
17:42:30.336077 IP 172.16.111.102 > 172.16.111.103: IP 10.10.200.10.123 > 10.10.100.10.37476: Flags [P.], seq 1:247, ack 81, win 502, options [nop,nop,TS val 1533882345 ecr 2882874243], length 246
17:42:30.336172 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [.], ack 247, win 506, options [nop,nop,TS val 2882874244 ecr 1533882345], length 0
17:42:30.336264 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [F.], seq 81, ack 247, win 506, options [nop,nop,TS val 2882874244 ecr 1533882345], length 0
17:42:30.336282 IP 172.16.111.102 > 172.16.111.103: IP 10.10.200.10.123 > 10.10.100.10.37476: Flags [F.], seq 247, ack 82, win 502, options [nop,nop,TS val 1533882346 ecr 2882874244], length 0
17:42:30.336387 IP 172.16.111.103 > 172.16.111.102: IP 10.10.100.10.37476 > 10.10.200.10.123: Flags [.], ack 248, win 506, options [nop,nop,TS val 2882874244 ecr 1533882346], length 0

```

抓包 ipip.1 端口, 已经是解包后的包:

```shell
tcpdump -i ipip.1 -nneevv -A icmp

tcpdump -i ipip.1 -nneevv -A port 123 -w ipip.pcap
tcpdump -i ipip.1 -nneevv -A port 123
root@xxx:~# tcpdump -i ipip.1 -n port 123
tcpdump: verbose output suppressed, use -v[v]... for full protocol decode
listening on tun2, link-type RAW (Raw IP), snapshot length 262144 bytes
17:43:44.522429 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [S], seq 745752260, win 64800, options [mss 1440,sackOK,TS val 2882948430 ecr 0,nop,wscale 7], length 0
17:43:44.522477 IP 10.10.200.10.123 > 10.10.100.10.47574: Flags [S.], seq 342785271, ack 745752261, win 64260, options [mss 1440,sackOK,TS val 1533956532 ecr 2882948430,nop,wscale 7], length 0
17:43:44.522615 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [.], ack 1, win 507, options [nop,nop,TS val 2882948430 ecr 1533956532], length 0
17:43:44.522648 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [P.], seq 1:81, ack 1, win 507, options [nop,nop,TS val 2882948430 ecr 1533956532], length 80
17:43:44.522660 IP 10.10.200.10.123 > 10.10.100.10.47574: Flags [.], ack 81, win 502, options [nop,nop,TS val 1533956532 ecr 2882948430], length 0
17:43:44.522787 IP 10.10.200.10.123 > 10.10.100.10.47574: Flags [P.], seq 1:247, ack 81, win 502, options [nop,nop,TS val 1533956532 ecr 2882948430], length 246
17:43:44.522842 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [.], ack 247, win 506, options [nop,nop,TS val 2882948430 ecr 1533956532], length 0
17:43:44.522999 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [F.], seq 81, ack 247, win 506, options [nop,nop,TS val 2882948431 ecr 1533956532], length 0
17:43:44.523016 IP 10.10.200.10.123 > 10.10.100.10.47574: Flags [F.], seq 247, ack 82, win 502, options [nop,nop,TS val 1533956532 ecr 2882948431], length 0
17:43:44.523092 IP 10.10.100.10.47574 > 10.10.200.10.123: Flags [.], ack 248, win 506, options [nop,nop,TS val 2882948431 ecr 1533956532], length 0

```

## 结论
icmp 报文(iphdr-1+icmp) 外层包装一个 outer iphdr-2，iphdr-1 里 srcIP/dstIP 是两个 ipip.1 vtep ip 地址，iphdr-2 里的 srcIP/dstIP 是两个 node ip 地址。
相比于 vxlan 把 original L2 frame 封装到 udp 报文里，即 outer ethhdr + outer iphdr + udphdr + vxlan hdr + L2，要简单很多。
