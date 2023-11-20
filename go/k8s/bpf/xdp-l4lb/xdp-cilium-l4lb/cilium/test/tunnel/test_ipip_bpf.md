





# tcpdump 抓包



```

# ipip_tunnel1 网卡里看到的还是 icmp 包，在 tc egress 做了 ipip 封包
root@xxx:~/tunnel# tcpdump -i ipip_tunnel1 -nneevv -A 
14:36:28.912542 ip: (tos 0x0, ttl 64, id 10643, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.200 > 10.1.1.100: ICMP echo request, id 21, seq 1, length 64
14:36:28.912583 ip: (tos 0x0, ttl 64, id 12000, offset 0, flags [none], proto ICMP (1), length 84)
    10.1.1.100 > 10.1.1.200: ICMP echo reply, id 21, seq 1, length 64

# 经过 ipip_tunnel1 后封包，抓包 ipip_veth1 看到 ipip 包
root@xxx:~/tunnel# tcpdump -i ipip_veth1 -nneevv -A proto 4
tcpdump: listening on ipip_veth1, link-type EN10MB (Ethernet), snapshot length 262144 bytes
14:24:17.891447 9e:38:1f:6e:38:32 > 4e:66:77:3a:3d:54, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 1466, offset 0, flags [none], proto IPIP (4), length 104)
    173.16.1.200 > 173.16.1.100: (tos 0x0, ttl 64, id 121, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.200 > 10.1.1.100: ICMP echo request, id 20, seq 1, length 64
14:24:17.891473 4e:66:77:3a:3d:54 > 9e:38:1f:6e:38:32, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 40915, offset 0, flags [DF], proto IPIP (4), length 104)
    173.16.1.100 > 173.16.1.200: (tos 0x0, ttl 64, id 50991, offset 0, flags [none], proto ICMP (1), length 84)
    10.1.1.100 > 10.1.1.200: ICMP echo reply, id 20, seq 1, length 64

# 抓包 ipip_ns0 里的 ipip_veth0，和抓包 ipip_veth1 一样，都是 ipip 包，两个网卡是一对 veth-pair
root@xxx:~/tunnel# ip netns exec ipip_ns0 tcpdump -i ipip_veth0 -nneevv -A
tcpdump: listening on ipip_veth0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
14:39:46.629199 9e:38:1f:6e:38:32 > 4e:66:77:3a:3d:54, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 60277, offset 0, flags [none], proto IPIP (4), length 104)
    173.16.1.200 > 173.16.1.100: (tos 0x0, ttl 64, id 14529, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.200 > 10.1.1.100: ICMP echo request, id 22, seq 1, length 64
14:39:46.629216 4e:66:77:3a:3d:54 > 9e:38:1f:6e:38:32, ethertype IPv4 (0x0800), length 118: (tos 0x0, ttl 64, id 44456, offset 0, flags [DF], proto IPIP (4), length 104)
    173.16.1.100 > 173.16.1.200: (tos 0x0, ttl 64, id 14251, offset 0, flags [none], proto ICMP (1), length 84)
    10.1.1.100 > 10.1.1.200: ICMP echo reply, id 22, seq 1, length 64

# 抓包 ipip 网卡 ipip_tunnel0，看到的是解包后的包
root@xxx:~/tunnel# ip netns exec ipip_ns0 tcpdump -i ipip_tunnel0 -nneevv -A
tcpdump: listening on ipip00_tunnel, link-type RAW (Raw IP), snapshot length 262144 bytes
14:40:27.692765 ip: (tos 0x0, ttl 64, id 22349, offset 0, flags [DF], proto ICMP (1), length 84)
    10.1.1.200 > 10.1.1.100: ICMP echo request, id 23, seq 1, length 64
14:40:27.692783 ip: (tos 0x0, ttl 64, id 20883, offset 0, flags [none], proto ICMP (1), length 84)
    10.1.1.100 > 10.1.1.200: ICMP echo reply, id 23, seq 1, length 64

```



