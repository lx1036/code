






# 抓包验证 ipip
参考自文章: https://blog.csdn.net/qq_22918243/article/details/130886641


https://developers.redhat.com/blog/2019/05/17/an-introduction-to-linux-virtual-interfaces-tunnels#ipip
ipip linux 源码: /root/linux-5.10.142/net/ipv4/ipip.c
ipip wiki: https://en.wikipedia.org/wiki/IP_in_IP
ipip shell 命令验证: https://blog.csdn.net/qq_22918243/article/details/130886641


两台 ecs 验证:

```shell
# nodeIP
ecs1 172.16.111.103
ecs2 172.16.111.102

# ipip 里 remote 和 local 用来匹配包，如果不想具体指定，可以设置为 remote any local any

# ecs1
ip tunnel add tun1 mode ipip remote 172.16.111.102 local 172.16.111.103
ip link set tun1 up
ip addr add 10.10.100.10 peer 10.10.200.10 dev tun1

# ecs2
ip tunnel add tun2 mode ipip remote 172.16.111.103 local 172.16.111.102
ip link set tun2 up
ip addr add 10.10.200.10 peer 10.10.100.10 dev tun2

root@xxx:~# ip addr show tun1
5: tun1@NONE: <POINTOPOINT,NOARP,UP,LOWER_UP> mtu 1480 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ipip 172.16.111.103 peer 172.16.111.102
    inet 10.10.100.10 peer 10.10.200.10/32 scope global tun1
       valid_lft forever preferred_lft forever
    inet6 fe80::5efe:ac10:6f67/64 scope link 
       valid_lft forever preferred_lft forever

root@xxx:~# ip addr show tun2
5: tun2@NONE: <POINTOPOINT,NOARP,UP,LOWER_UP> mtu 1480 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ipip 172.16.111.102 peer 172.16.111.103
    inet 10.10.200.10 peer 10.10.100.10/32 scope global tun2
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

必须抓包 eth0 端口，而不是 tun2 端口:

```shell
# 指定 ip src 和 ip dst
tcpdump -i eth0 -nneevv -A ip src 172.16.111.103 and ip dst 172.16.111.102

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


抓包 tun2 端口, 已经是解包后的包:


```shell
tcpdump  -i tun2 -nneevv -A port 123 -w ipip.pcap
tcpdump  -i tun2 -nneevv -A port 123

root@xxx:~# tcpdump  -i tun2 -n port 123
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

