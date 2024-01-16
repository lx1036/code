
#!/bin/bash

############ 1.点对点 vxlan demo 测试(验证通过) ############
# [Linux 下实践 VxLAN](https://mp.weixin.qq.com/s/bsoWU2WC6SxPgseNxwG9cw)

# 构造两个 ns 来模拟两个 vm
# ns1(eth0, vxlan1) <---> ns2(eth0, vxlan1), 两个 vxlan1 在一个网段，使用 ns1/ns2 模拟两个 vm

ip netns add vxlan-ns1
ip netns add vxlan-ns2
ip link add vxlan-veth0 type veth peer name vxlan-veth1
ip link set vxlan-veth0 netns vxlan-ns1
ip link set vxlan-veth1 netns vxlan-ns2
ip netns exec vxlan-ns1 ip link set dev vxlan-veth0 name eth0
ip netns exec vxlan-ns2 ip link set dev vxlan-veth1 name eth0
ip netns exec vxlan-ns1 ip link set eth0 up
ip netns exec vxlan-ns2 ip link set eth0 up
ip netns exec vxlan-ns1 ip addr add 172.31.0.106/24 dev eth0
ip netns exec vxlan-ns2 ip addr add 172.31.0.107/24 dev eth0
# 连通性
ip netns exec vxlan-ns1 ping -c 3 172.31.0.107
ip netns exec vxlan-ns2 ping -c 3 172.31.0.106

# "remote 172.31.0.107 dstport 4789 dev eth0" 这里是对方的网络数据
# 在构造 Outer Mac/IP Header 时需要填充 Outer_Mac_Header(node1_src_mac/node2_dst_mac), Outer_IP_Header(node1_src_ip/node2_dst_ip)
ip netns exec vxlan-ns1 ip link add vxlan1 type vxlan id 100 remote 172.31.0.107 dstport 4789 dev eth0
ip netns exec vxlan-ns2 ip link add vxlan1 type vxlan id 100 remote 172.31.0.106 dstport 4789 dev eth0
ip netns exec vxlan-ns1 ip link set vxlan1 up
ip netns exec vxlan-ns2 ip link set vxlan1 up
ip netns exec vxlan-ns1 ip addr add 10.0.0.106/24 dev vxlan1
ip netns exec vxlan-ns2 ip addr add 10.0.0.107/24 dev vxlan1

ip netns exec vxlan-ns1 ip -d addr show vxlan1
# 注意这里的 "vxlan id 100 remote 172.31.0.107 dev eth0 srcport 0 0 dstport 4789"
4: vxlan1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ether 6a:16:55:92:a3:e5 brd ff:ff:ff:ff:ff:ff promiscuity 0
    vxlan id 100 remote 172.31.0.107 dev eth0 srcport 0 0 dstport 4789 ageing 300 udpcsum noudp6zerocsumtx noudp6zerocsumrx numtxqueues 1 numrxqueues 1 gso_max_size 65536 gso_max_segs 65535
    inet 10.0.0.106/24 scope global vxlan1
       valid_lft forever preferred_lft forever
    inet6 fe80::6816:55ff:fe92:a3e5/64 scope link
       valid_lft forever preferred_lft forever
ip netns exec vxlan-ns2 ip -d addr show vxlan1
# 查看fdb表
ip netns exec vxlan-ns1 bridge fdb show
ip netns exec vxlan-ns2 bridge fdb show
# 连通性
ip netns exec vxlan-ns1 ping -c 3 10.0.0.107
ip netns exec vxlan-ns2 ping -c 3 10.0.0.106


# 抓包
ip netns exec vxlan-ns2 tcpdump -i eth0 -nneevv -A -w vxlan-ns2-eth0.pcap # 抓包 icmp 和 arp，原始 icmp/arp 报文 在 UDP Data 里
ip netns exec vxlan-ns2 tcpdump -i vxlan1 -nneevv -A -w vxlan-ns2-vxlan1.pcap # 已经解包后的数据报文，即 UDP Data 里的报文
ip netns exec vxlan-ns1 ping -c 3 10.0.0.107

ip netns exec vxlan-ns1 tcpdump -i eth0 -nneevv -A -w vxlan-ns1-eth0.pcap
ip netns exec vxlan-ns1 tcpdump -i vxlan1 -nneevv -A -w vxlan-ns1-vxlan1.pcap
ip netns exec vxlan-ns2 ping -c 3 10.0.0.106

########################################################################



############ 2.跨节点 pod 通信的 vxlan demo 测试(验证通过) ############
# [flannel vxlan 模拟](https://mp.weixin.qq.com/s/uPbKZe2NBAwZRuWfiORrqw) https://juejin.cn/post/6994825163757846565

# node1{eth0, flannel.1, cni0[pod1(eth0<->veth1)]}, flannel.1/cni0/pod1 在一个网段，如 cni0 10.244.0.1/24, flannel.1 10.244.0.0/32, pod1 10.244.0.20/32
# <-------->
# node2{eth0, flannel.1, cni0[pod2(eth0<->veth1)]}, flannel.1/cni0/pod2 在一个网段，如 cni0 10.244.1.1/24, flannel.1 10.244.1.0/32, pod1 10.244.1.20/32

# 验证: pod1 <---> pod2 连通性

# 没法使用两个 ns 来验证，因为还得有 pod，只能两个 vm 来验证

# INFO: 验证 flannel vxlan 模式

# Node1 10.244.0.0/24
ip netns add vxlan-ns1
ip netns add vxlan-ns2
ip link add dev cni0 type bridge
ip link set cni0 up
ip link add pod1-veth0 type veth peer name pod1-veth1
ip link set pod1-veth1 up
ip link add pod2-veth0 type veth peer name pod2-veth1
ip link set pod2-veth1 up
ip link set pod1-veth1 master cni0
ip link set pod2-veth1 master cni0
ip link set pod1-veth0 netns vxlan-ns1
ip link set pod2-veth0 netns vxlan-ns2
ip netns exec vxlan-ns1 ip link set dev pod1-veth0 name eth0
ip netns exec vxlan-ns2 ip link set dev pod2-veth0 name eth0
ip netns exec vxlan-ns1 ip link set eth0 up
ip netns exec vxlan-ns1 ip link set lo up
ip netns exec vxlan-ns2 ip link set eth0 up
ip netns exec vxlan-ns2 ip link set lo up
ip addr add 10.244.0.1/24 dev cni0
ip netns exec vxlan-ns1 ip addr add 10.244.0.20/24 dev eth0
ip netns exec vxlan-ns2 ip addr add 10.244.0.21/24 dev eth0
ip netns exec vxlan-ns1 ip route add default dev eth0
ip netns exec vxlan-ns2 ip route add default dev eth0
# 同节点 pod 通过 cni0 bridge 互通
ip netns exec vxlan-ns1 ping -c 3 10.244.0.21
ip netns exec vxlan-ns2 ping -c 3 10.244.0.20

ip link add flannel.1 type vxlan id 100 remote 172.16.10.2 dstport 4789 dev eth0
ip link set flannel.1 up
ip addr add 10.244.0.0/32 dev flannel.1
# 如果缺少这一步，没法实现 arp 报文由 cni0 交给 flannel.1 处理，从而封装的 arp 报文(`tcpdump -i eth0 -nneevv -A udp and port 4789`) 没有应答
ip link set flannel.1 master cni0
ip route add 10.244.1.0/24 via 10.244.1.0 dev flannel.1
# 验证跨节点 pod 连通性
ip netns exec vxlan-ns1 ping -c 3 10.244.1.20
ip netns exec vxlan-ns1 ping -c 3 10.244.1.21

# Node2 10.244.1.0/24
ip netns add vxlan-ns1
ip netns add vxlan-ns2
ip link add dev cni0 type bridge
ip link set cni0 up
ip link add pod1-veth0 type veth peer name pod1-veth1
ip link set pod1-veth1 up
ip link add pod2-veth0 type veth peer name pod2-veth1
ip link set pod2-veth1 up
ip link set pod1-veth1 master cni0
ip link set pod2-veth1 master cni0
ip link set pod1-veth0 netns vxlan-ns1
ip link set pod2-veth0 netns vxlan-ns2
ip netns exec vxlan-ns1 ip link set dev pod1-veth0 name eth0
ip netns exec vxlan-ns2 ip link set dev pod2-veth0 name eth0
ip netns exec vxlan-ns1 ip link set eth0 up
ip netns exec vxlan-ns1 ip link set lo up
ip netns exec vxlan-ns2 ip link set eth0 up
ip netns exec vxlan-ns2 ip link set lo up
ip addr add 10.244.1.1/24 dev cni0
ip netns exec vxlan-ns1 ip addr add 10.244.1.20/24 dev eth0
ip netns exec vxlan-ns2 ip addr add 10.244.1.21/24 dev eth0
ip netns exec vxlan-ns1 ip route add default dev eth0
ip netns exec vxlan-ns2 ip route add default dev eth0
# 同节点 pod 通过 cni0 bridge 互通
ip netns exec vxlan-ns1 ping -c 3 10.244.1.21
ip netns exec vxlan-ns2 ping -c 3 10.244.1.20

ip link add flannel.1 type vxlan id 100 remote 172.16.10.3 dstport 4789 dev eth0
ip link set flannel.1 up
ip addr add 10.244.1.0/32 dev flannel.1
# 如果缺少这一步，没法实现 arp 报文由 cni0 交给 flannel.1 处理，从而封装的 arp 报文(`tcpdump -i eth0 -nneevv -A udp and port 4789`) 没有应答
ip link set flannel.1 master cni0
ip route add 10.244.0.0/24 dev flannel.1
# 验证跨节点 pod 连通性
tcpdump -i eth0 -nneevv -A udp and port 4789
ip netns exec vxlan-ns1 ping -c 3 10.244.0.20
ip netns exec vxlan-ns1 ping -c 3 10.244.0.21

cleanup() {
  ip netns del vxlan-ns1
  ip netns del vxlan-ns2
  ip link del cni0
  ip link del flannel.1
}


########################################################################



############ 3.跨节点 pod 通信的 vxlan demo 测试之手动更新FDB表来实现VXLAN通信(验证通过) ############
# [动态维护 FDB 表项实现 VxLAN 通信](https://mp.weixin.qq.com/s/PDh_6JLr_yV5gfbHZV5NIw)

