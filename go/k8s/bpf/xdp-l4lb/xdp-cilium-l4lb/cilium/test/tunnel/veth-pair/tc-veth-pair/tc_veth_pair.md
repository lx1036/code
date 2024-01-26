

# tc ingress/egress

```md
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tc_neigh_fib.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tc_neigh.c
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tc_peer.c
/root/linux-5.10.142/tools/testing/selftests/bpf/test_tc_redirect.sh
```


# 抓包
通过抓包发现，网络加速跳过了 veth_src_fwd 和 veth_dst_fwd 两个网卡 netfilter，等同于 veth_src 和 veth_dst 直接互联。

```shell
ping -c 1 173.16.2.100

tcpdump -i veth_src_fwd -nneevv
4e:83:75:4a:ef:64 > 7e:73:7d:99:78:21, ethertype IPv4 (0x0800), length 98: (tos 0x0, ttl 64, id 51934, offset 0, flags [DF], proto ICMP (1), length 84)
    173.16.1.100 > 173.16.2.100: ICMP echo request, id 12772, seq 1, length 64

tcpdump -i veth_dst_fwd -nneevv
32:c9:01:d5:5d:76 > 62:f9:63:9d:cb:72, ethertype IPv4 (0x0800), length 98: (tos 0x0, ttl 64, id 60063, offset 0, flags [none], proto ICMP (1), length 84)
    173.16.2.100 > 173.16.1.100: ICMP echo reply, id 52763, seq 1, length 64    
```

