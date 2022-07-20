

# Flannel
主要实现两个功能:
* 为每个 node 分配 subnet，容器将自动从该子网中获取 IP 地址
* 当有 node 加入到网络中时，为每个 node 增加路由配置


# Vxlan 
与 VLAN 对比：
1. VXLAN TAG 有 24 位。能提供16777216个逻辑网络标识符，VXLAN的标识符称为VNI(VXLAN Network Identifier)。
2. VLAN只能应用在一个二层网络中，而VXLAN通过将原始二层以太网帧封装在IP协议包中，在IP基础网络之上构建overlay的逻辑大二层网络


FDB: Forward Database, 转发表，它是二层交换设备用来数据通信的，它里面包含MAC 地址 、VLAN 号、端口号和一些标志信息。主要功能就是：根据 mac 查找 vtep ip。
VNI: VXLAN Network Identifier
VTEP: Vxlan Tunnel Endpoint


## Vxlan 通信过程
对于处于同一个VXLAN的两台虚拟终端，其通信过程可以概括为如下的步骤：
1. 发送方向接收方发送数据帧，帧中包含了发送方和接收方的虚拟MAC地址
2. 发送方连接的VTEP节点收到了数据帧，通过查找发送方所在的VXLAN以及接收方所连接的VTEP节点，将该报文添加VXLAN首部、外部UDP首部、外部IP首部后，发送给目的VTEP节点
3. 报文经过物理网络传输到达目的VTEP节点
4. 目的VTEP节点接收到报文后，拆除报文的外部IP首部和外部UDP首部，检查报文的VNI以及内部数据帧的目的MAC地址，确认接收方与本VTEP节点相连后，拆除VXLAN首部，将内部数据帧交付给接收方
5. 接收方收到数据帧，传输完成

通过以上的步骤可以看出：VXLAN的实现细节以及通信过程对于处于VXLAN中的发送方和接收方是不可见的，基于发送方和接收方的视角，其通信过程和二者真实处于同一链路层网络中的情况完全相同。

## 验证 Vxlan 通信过程

```shell
# 安装 bridge 命令
yum install -y iproute iproute-devel iproute-tc

# 自学习模式(没有验证通过)
# node1
ip link add name vxlan-br0 type bridge
ip addr add 100.1.1.2/24 dev vxlan-br0
ip link set up vxlan-br0
ip addr show vxlan-br0
ip link add vxlan0 type vxlan id 1 group 239.1.1.1 dev eth2 dstport 4789 # 这里 eth2 是出口网卡
# 把 br-veth0 搭到 vxlan-br0 网桥上
ip link set dev vxlan0 master vxlan-br0 # brctl addif vxlan-br0 vxlan0
ip link del vxlan0
ip link del vxlan-br0
# node2
ip link add name vxlan-br0 type bridge
ip addr add 100.1.1.3/24 dev vxlan-br0
ip link set up vxlan-br0
ip addr show vxlan-br0
ip link del vxlan0
ip link del vxlan-br0


# 手动更新FDB表来实现VXLAN通信(验证通过)
# @see http://just4coding.com/2020/04/20/vxlan-fdb/
# node1
sysctl -w net.ipv4.ip_forward=1
ip netns add ns1
ip link add veth1 type veth peer name eth0 netns ns1
ip netns exec ns1 ip link set eth0 up
ip netns exec ns1 ip link set lo up
ip netns exec ns1 ip addr add 3.3.3.3/24 dev eth0
ip link set up dev veth1
ip netns exec ns1 ip addr
ip addr show dev veth1
ip link add br1 type bridge
ip link set br1 up
ip link set veth1 master br1 # 把 veth1 搭到 br1 网桥上 # brctl addif vxlan-br0 vxlan0
# 这里 10.206.67.96 是 node2 的 nodeIP, 指定了nolearning来禁用源地址学习
ip link add vxlan100 type vxlan id 100 dstport 4789 local 10.206.67.96 nolearning
ip link set vxlan100 master br1
ip link set up vxlan100
ip addr show dev vxlan100
# node2
sysctl -w net.ipv4.ip_forward=1
ip netns add ns1
ip link add veth1 type veth peer name eth0 netns ns1
ip netns exec ns1 ip link set eth0 up
ip netns exec ns1 ip link set lo up
ip netns exec ns1 ip addr add 3.3.3.4/24 dev eth0
ip link set up dev veth1
ip netns exec ns1 ip addr
ip addr show dev veth1
ip link add br1 type bridge
ip link set br1 up
ip link set veth1 master br1
# 这里 10.206.67.15 是 node1 的 nodeIP, 指定了nolearning来禁用源地址学习
ip link add vxlan100 type vxlan id 100 dstport 4789 local 10.206.67.15 nolearning
ip link set vxlan100 master br1
ip link set up vxlan100
ip addr show dev vxlan100

# 验证跨节点的同一网络平面内 3.3.3.0/24 是否可达，即是否实现了 overlay 大二层
# node1
ip netns exec ns1 ping -c 2 3.3.3.4
ip netns exec ns1 ip neigh # 注意：这里会 ARP 学习到 3.3.3.4 容器的 mac 地址
# node2
ip netns exec ns1 ping -c 2 3.3.3.3

# 以上禁止了 nolearning，还可以添加 Linux VXLAN设备 proxy 参数开启ARP代答，注意 proxy 参数
# 两个 node 上分别操作：
[root@docker12 liuxiang3]# ip link del vxlan100
[root@docker12 liuxiang3]# ip link add vxlan100 type vxlan id 100 dstport 4789 local 10.206.67.96 nolearning proxy
[root@docker12 liuxiang3]# ip link set vxlan100 master br1
[root@docker12 liuxiang3]# ip link set up vxlan100
[root@docker12 liuxiang3]# bridge fdb show brport vxlan100
[root@docker12 liuxiang3]# bridge fdb append 32:39:e3:57:89:e7 dev vxlan100 dst 10.206.67.15 # 10.206.67.15 是另一台 node 的 nodeIP
[root@docker12 liuxiang3]# bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 10.206.67.15
[root@docker12 liuxiang3]# ip neighbor add 3.3.3.4 lladdr 32:39:e3:57:89:e7 dev vxlan100
[root@docker12 liuxiang3]# ip neighbor show dev vxlan100
[root@docker12 liuxiang3]# ip netns exec ns1 ping -c 2 3.3.3.4

[root@docker03 liuxiang3]# ip link del vxlan100
[root@docker03 liuxiang3]# ip link add vxlan100 type vxlan id 100 dstport 4789 local 10.206.67.15 nolearning proxy
[root@docker03 liuxiang3]# ip link set vxlan100 master br1
[root@docker03 liuxiang3]# ip link set up vxlan100
[root@docker03 liuxiang3]# bridge fdb append aa:bf:b9:29:63:12 dev vxlan100 dst 10.206.67.96
[root@docker03 liuxiang3]# bridge fdb append 00:00:00:00:00:00 dev vxlan100 dst 10.206.67.96
[root@docker03 liuxiang3]# ip neighbor add 3.3.3.3 lladdr aa:bf:b9:29:63:12 dev vxlan100
[root@docker03 liuxiang3]# ip neighbor show dev vxlan100
[root@docker03 liuxiang3]# ip netns exec ns1 ip neigh flush all # 清空 ARP
[root@docker03 liuxiang3]# ip netns exec ns1 ip neigh # 空的 
[root@docker03 liuxiang3]# ip netns exec ns1 ping -c 2 3.3.3.3 # 然后会学习到 3.3.3.3 的 mac 地址

```




# 参考文献
**[Kubernetes flannel网络分析](http://just4coding.com/2021/11/03/flannel/)** , 这个很重要！！！

**[动态维护FDB表项实现VXLAN通信](http://just4coding.com/2020/04/20/vxlan-fdb/)** , 这个很重要！！！

**[VXLAN原理介绍与实例分析](http://just4coding.com/2017/05/21/vxlan/)**

**[VxLAN 与 Bridge、Namespace基础](https://mp.weixin.qq.com/s/JYp36vfX8r0l7VlCGMK8kA)** , 这个很重要！！！
