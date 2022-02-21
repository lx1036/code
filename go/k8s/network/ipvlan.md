
# IPVlan 和 MacVlan
相比于 linux-bridge 模式，不需要一个 bridge，更加简便。


# 如何连通两个 ns 之间的容器？

terway ipvlan L2: https://github.com/AliyunContainerService/terway/blob/main/docs/design.md#ipvlan-l2
ipvlan: 从一个网卡虚拟出多个网卡，这些网卡拥有相同的mac地址，但是有不同的IP地址。

需要注意的地方：DHCP 协议分配 ip 的时候一般会用 mac 地址作为机器的标识。这个情况下，客户端动态获取 ip 的时候需要配置唯一的 ClientID 字段，
并且 DHCP server 也要正确配置使用该字段作为机器标识，而不是使用 mac 地址

## 两种模式
ipvlan 有两种不同的模式：L2 和 L3。

### L2 模式

```shell
ip netns add ns0
ip netns add ns1

ip link add link eth0 ipvl0 type ipvlan mode l2
ip link add link eth0 ipvl1 type ipvlan mode l2

ip link set dev ipvl0 netns ns0
ip link set dev ipvl1 netns ns1

# For ns0
ip netns exec ns0 bash
ip link set dev ipvl0 up
ip link set dev lo up
ip -4 addr add 127.0.0.1 dev lo
ip -4 addr add $IPADDR dev ipvl0
ip -4 route add default via $ROUTER dev ipvl0
# For ns1
ip netns exec ns1 bash
ip link set dev ipvl1 up
ip link set dev lo up
ip -4 addr add 127.0.0.1 dev lo
ip -4 addr add $IPADDR dev ipvl1
ip -4 route add default via $ROUTER dev ipvl1
```

### L3 模式

```shell
# 测试使用 IPVlan L3 模式下两个 ns 下的容器网络互通
sudo ip netns add net1
sudo ip netns add net2
# 分别创建 ipvlan 网卡，父网卡是 eth0
sudo ip link add ipv1 link eth0 type ipvlan mode l3
sudo ip link add ipv2 link eth0 type ipvlan mode l3
# 移动网卡到对应的 ns
sudo ip link set ipv1 netns net1
sudo ip link set ipv2 netns net2
sudo ip netns exec net1 ip link set ipv1 up # ip netns exec net1 bash
sudo ip netns exec net2 ip link set ipv2 up
# 配置 ip 地址和默认路由
sudo ip netns exec net1 ip addr add 10.0.1.10/24 dev ipv1
sudo ip netns exec net2 ip addr add 10.0.2.10/24 dev ipv2
sudo ip netns exec net1 ip route add default dev ipv1
sudo ip netns exec net2 ip route add default dev ipv2
# 测试两个 ns 的容器连通性
sudo ip netns exec net1 ping -c 3 10.0.2.10
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
