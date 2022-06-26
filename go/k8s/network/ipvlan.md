
# IPVlan 和 MacVlan
相比于 linux-bridge 模式，不需要一个 bridge，更加简便。


# 如何连通两个 ns 之间的容器？

terway ipvlan L2: https://github.com/AliyunContainerService/terway/blob/main/docs/design.md#ipvlan-l2
ipvlan: 从一个网卡虚拟出多个网卡，这些网卡拥有相同的mac地址，但是有不同的IP地址。

需要注意的地方：DHCP 协议分配 ip 的时候一般会用 mac 地址作为机器的标识。这个情况下，客户端动态获取 ip 的时候需要配置唯一的 ClientID 字段，
并且 DHCP server 也要正确配置使用该字段作为机器标识，而不是使用 mac 地址

## 两种模式
ipvlan 有两种不同的模式：L2(二层交换机)和 L3(三层交换机)。
二层交换机 和 三层交换机 主要的区别就是：二层交换机通过 mac 地址(arp 协议)来决定下一跳；三层交换机通过 ip 决定下一跳，下一跳 ip 通过查询路由表路由。
**[二、三层交换机之间到底有什么区别](https://mp.weixin.qq.com/s/U_-fjMPvh1W4_c1ao34YAg)**

### L2 模式

```shell
ip link add dummy-ipvlan-l2 type dummy
ip link set dummy-ipvlan-l2 up
ip netns add net-ipvlan-l2-1
ip netns add net-ipvlan-l2-2
ip link add ipv1 link dummy-ipvlan-l2 type ipvlan mode l2
ip link add ipv2 link dummy-ipvlan-l2 type ipvlan mode l2
ip link add ipv3 link dummy-ipvlan-l2 type ipvlan mode l2
ip link set ipv1 netns net-ipvlan-l2-1
ip link set ipv2 netns net-ipvlan-l2-2
ip link set ipv3 netns net-ipvlan-l2-1
ip netns exec net-ipvlan-l2-1 ip link set ipv1 up
ip netns exec net-ipvlan-l2-2 ip link set ipv2 up
ip netns exec net-ipvlan-l2-1 ip link set ipv3 up
ip netns exec net-ipvlan-l2-1 ip addr add 200.1.1.10/24 dev ipv1
ip netns exec net-ipvlan-l2-2 ip addr add 200.1.2.10/24 dev ipv2
ip netns exec net-ipvlan-l2-1 ip addr add 200.2.1.10/32 dev ipv3
ip netns exec net-ipvlan-l2-1 ip route add default dev ipv1
ip netns exec net-ipvlan-l2-2 ip route add default dev ipv2
ip netns exec net-ipvlan-l2-1 ping -c 3 200.1.2.10
ip netns exec net-ipvlan-l2-2 ping -c 3 200.1.1.10
ip netns exec net-ipvlan-l2-2 ping -c 3 200.2.1.10
```

### L3 模式

```shell
# 测试使用 IPVlan L3 模式下两个 net namespace 下的容器网络互通

# 父网卡是自建的 dummy type 虚拟网卡
ip link add dummy-ipvlan-l3 type dummy
ip link set dummy-ipvlan-l3 up
ip netns add net-ipvlan-l3-3
ip netns add net-ipvlan-l3-4
ip link add ipv1 link dummy-ipvlan-l3 type ipvlan mode l3
ip link add ipv2 link dummy-ipvlan-l3 type ipvlan mode l3
ip link add ipv3 link dummy-ipvlan-l3 type ipvlan mode l3
# 移动网卡到对应的 ns
ip link set ipv1 netns net-ipvlan-l3-3
ip link set ipv2 netns net-ipvlan-l3-4
ip link set ipv3 netns net-ipvlan-l3-3
ip netns exec net-ipvlan-l3-3 ip link set ipv1 up
ip netns exec net-ipvlan-l3-4 ip link set ipv2 up
ip netns exec net-ipvlan-l3-3 ip link set ipv3 up
# 配置 ip 地址和默认路由
ip netns exec net-ipvlan-l3-3 ip addr add 200.0.1.10/24 dev ipv1
ip netns exec net-ipvlan-l3-4 ip addr add 200.0.2.10/24 dev ipv2
ip netns exec net-ipvlan-l3-3 ip addr add 210.0.1.10/32 dev ipv3
ip netns exec net-ipvlan-l3-3 ip route add default dev ipv1
ip netns exec net-ipvlan-l3-4 ip route add default dev ipv2
# 测试两个 ns 的容器连通性
ip netns exec net-ipvlan-l3-3 ping -c 3 200.0.2.10
ip netns exec net-ipvlan-l3-4 ping -c 3 200.0.1.10
ip netns exec net-ipvlan-l3-4 ping -c 3 210.0.1.10
```

问题：但是在 host net namespace 下无法 ping/curl 通 200.0.2.10 等 ipvlan 网卡，以及 service ip？

```shell
# veth pair 打通容器网络
# 可以看这个，已经经过验证
ip link add veth-test-2 type veth peer name veth-test-3
ip netns add net-veth-2
ip link set veth-test-2 netns net-veth-2
ip link set veth-test-2 up
ip netns exec net-veth-2 ip link set veth-test-2 up
ip netns exec net-veth-2 ip route add 169.254.1.1 dev veth-test-2
ip netns exec net-veth-2 ip route add default via 169.254.1.1 dev veth-test-2
ip netns exec net-veth-2 ip neigh add 169.254.1.1 dev veth-test-2 lladdr ee:ee:ee:ee:ee:ee
ip link set addr ee:ee:ee:ee:ee:ee veth-test-3
ip netns exec net-veth-2 ip addr add 100.162.253.162 dev veth-test-2 # 100.162.253.162 随便写的 pod ip
ip route add 100.162.253.162 dev veth-test-3
ip netns exec net-veth-2 curl -I 192.168.246.174 # 192.168.246.174 为 service ip
```


# IPVlan in Cilium
https://docs.cilium.io/en/v1.9/concepts/ebpf/lifeofapacket/#veth-based-versus-ipvlan-based-datapath :
ipvlan 相比于 veth-pair 模式优点在于，从 host namespace 到 container namespace，packets 不需要两次遍历 linux 内核网络协议栈。

ipvlan in Cilium 最新版本已经废弃，使用 eBPF 已经有了更好性能，且社区也不感兴趣：
https://docs.cilium.io/en/stable/operations/upgrade/#deprecated-options


# 参考文献
**[ipvlan 内核文档](https://www.kernel.org/doc/Documentation/networking/ipvlan.txt)**
**[网络虚拟化-ipvlan](https://cizixs.com/2017/02/17/network-virtualization-ipvlan/)**
**[网络虚拟化-macvlan](https://cizixs.com/2017/02/14/network-virtualization-macvlan/)**
**[Docker: Use IPvlan networks](https://docs.docker.com/network/ipvlan/)**
**[Docker: Use macvlan networks](https://docs.docker.com/network/macvlan/)**
**[书籍：Kubernetes 网络权威指南 1.8-1.9 小节]**
**[Terway IPVlan in Cilium](https://github.com/cilium/cilium/pull/10251)**
