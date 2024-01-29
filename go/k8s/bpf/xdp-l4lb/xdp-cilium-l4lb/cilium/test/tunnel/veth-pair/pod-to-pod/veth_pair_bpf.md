
# 目的
验证 tc ebpf 加速了通节点 pod 之间的网络包的转发过程。


# 抓包

```shell
ip netns exec pod1_ns ping -c 1 100.0.1.2

tcpdump -i pod1_veth -nneevv -A
00:02:27.958613 96:27:df:ff:43:19 > b2:6b:98:9e:2d:34, ethertype IPv4 (0x0800), length 98: (tos 0x0, ttl 64, id 23377, offset 0, flags [DF], proto ICMP (1), length 84)
    100.0.1.1 > 100.0.1.2: ICMP echo request, id 23397, seq 1, length 64
00:02:33.059357 96:27:df:ff:43:19 > b2:6b:98:9e:2d:34, ethertype ARP (0x0806), length 42: Ethernet (len 6), IPv4 (len 4), Request who-has 100.0.1.2 tell 100.0.1.1, length 28
00:02:33.059379 b2:6b:98:9e:2d:34 > 96:27:df:ff:43:19, ethertype ARP (0x0806), length 42: Ethernet (len 6), IPv4 (len 4), Reply 100.0.1.2 is-at b2:6b:98:9e:2d:34, length 28

tcpdump -i pod2_veth -nneevv -A
00:02:27.958642 f2:c6:a8:e2:b4:d1 > f6:aa:db:df:e4:bd, ethertype IPv4 (0x0800), length 98: (tos 0x0, ttl 64, id 14765, offset 0, flags [none], proto ICMP (1), length 84)
    100.0.1.2 > 100.0.1.1: ICMP echo reply, id 23397, seq 1, length 64
00:02:33.059348 f2:c6:a8:e2:b4:d1 > f6:aa:db:df:e4:bd, ethertype ARP (0x0806), length 42: Ethernet (len 6), IPv4 (len 4), Request who-has 100.0.1.1 tell 100.0.1.2, length 28
00:02:33.059373 f6:aa:db:df:e4:bd > f2:c6:a8:e2:b4:d1, ethertype ARP (0x0806), length 42: Ethernet (len 6), IPv4 (len 4), Reply 100.0.1.1 is-at f6:aa:db:df:e4:bd, length 28

```

通过抓包，得出 ebpf 加速了包的转发过程，相比于没有 ebpf, 减少了 pod1_veth 和 pod2_veth netfilter 内核协议栈:
```md
eth0(pod1_ns) -> tcpdump 抓包, pod1_veth(tc ingress) -> bpf_redirect_peer -> eth0(pod2_ns) -> 回包
-> tcpdump 抓包, pod2_veth(tc ingress) -> bpf_redirect_peer -> eth0(pod1_ns)
```

