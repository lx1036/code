
# Ingress是七层负载均衡，Service是四层负载均衡
四篇不错的技术文档：
**[Kubernetes 网络概念及策略控制](https://mp.weixin.qq.com/s/kjOAlKTwaMZzVOiSuJE6fQ)**:
k8s对于网络实现的约束约法三章：
* 第一条：任意两个 Pod 之间其实是可以直接通信的，无需经过显式地使用 NAT(Network Address Transfer) 来接收数据和地址的转换；
* 第二条：Node 与 Pod 之间是可以直接通信的，无需使用明显的地址转换；
* 第三条：Pod 看到自己的 IP 跟别人看见它所用的 IP 是一样的，中间不能经过转换。

想要设计一个容器网络，需要实现四大目标：
* 外部世界和 service 之间是怎么通信的？就是有一个互联网或者是公司外部的一个用户，怎么用到 service？
* service 如何与它后端的 pod 通讯？
* pod 和 pod 之间调用是怎么做到通信的？
* pod 内部容器与容器之间的通信？

根据容器网络与宿主网络寄生关系分为两种网络：**Underlay** 和 **Overlay**
* Underlay: **它与 Host 网络是同层的**。从外在可见的一个特征就是它是不是使用了 Host 网络同样的网段、输入输出基础设备、容器的 IP 地址是不是需要与 Host 网络取得协同（来自同一个中心分配或统一划分）
* Overlay: Overlay 不一样的地方就在于它并不需要从 Host 网络的 IPM(IP Manager) 的管理的组件去申请 IP，一般来说，它只需要跟 Host 网络不冲突，这个 IP 可以自由分配的。

### Network Namespace
Network Namespace 是实现*网络虚拟化*的内核基础，创建了**隔离的网络空间**:
* 拥有独立的网络设备(lo[loop back],veth pair等虚拟设备/物理网卡)，l
* 独立的协议栈，IP地址和路由表
* iptables 规则
* 独立的 ipvs等

Pod 与 Network Namespace 的关系，每个 pod 都有独立的网络空间:
![pod-netns](https://mmbiz.qpic.cn/mmbiz_png/yvBJb5Iiafvlb8OibYd4dhFaNUPC2ACB78w7ib880uV985T2DagSYxTPo3op8dNknKTeGKMYRccOmVdPyVmb3TB0w/640?wx_fmt=png&tp=webp&wxfrom=5&wx_lazy=1&wx_co=1)


#### Network Namespace 详解
(1)创建一个 network namespace
```shell script
# 会在 ./var/run/netns 目录下生成一个挂载点 netns1，就算没有进程在该网络里运行也可以存活。
ip netns list
ip netns add/delete netns1
ip netns exec netns1 ip link list # 进入 netns1 并 ip link list 查询
```
通过`ip netns exec netns1 ip link list`查询 netns1 下网卡信息，只有一个本地回还设备 lo:
```markdown
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN mode DEFAULT group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
```
这个回环设备状态还是 DOWN 的，ping 不通，需要打开：
```shell script
ip netns exec netns1 ping 127.0.0.1
# connect: Network is unreachable
ip netns exec netns1 ip link set dev lo up
ip netns exec netns1 ping -c 2 -w 2 127.0.0.1
# 2 packets transmitted, 2 received, 0% packet loss, time 1022ms
# 本地回还可以走通，但是没法与外界通信(与宿主主机上的网卡)，就需要在 netns1 内创建一对虚拟的以太网卡，即 veth pair。
```

查 root namespace 宿主机上网络设备，有网卡 eth0，docker0：
```shell script
ip link list
#1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
#    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
#2: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP mode DEFAULT group default qlen 1000
#    link/ether 00:16:3e:16:72:0e brd ff:ff:ff:ff:ff:ff
#3: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN mode DEFAULT group default
#    link/ether 02:42:97:f9:3e:1b brd ff:ff:ff:ff:ff:ff
```

(2)创建虚拟的以太网卡 veth(virtual ethernet) pair
```shell script
ip link add veth0 type veth peer name veth1 # 创建一对虚拟以太网卡 veth0 veth1，默认情况下都会在宿主机的 root network namespace
ip link list

# 把 veth0/veth1 up 起来
ip link set dev veth0 up
ip link set dev veth1 up

ip link set veth1 netns netns1 # 把 veth1 移到 netns1 network namespace
ip link list
 
# 把 veth0/veth1 up 起来，并绑定对应的IP地址，这样 veth pair 两头都可以 ping 通：
ip netns exec netns1 ifconfig veth1 10.1.1.1/24 up
ifconfig veth0 10.1.1.2/24 up
ping -c 2 -w 2 10.1.1.1
ip netns exec netns1 ping -c 2 -w 2 10.1.1.2 # 在 netns1 内 ping 主机上的虚拟网卡
```

不同 namespace 之间的路由表和防火墙规则也是隔离的：
```shell script
ip netns exec netns1 route
ip netns exec netns1 iptables -L

# 把 netns1 的虚拟网卡 veth1 移动到 pid=1 进程所在的 network namespace，即 root network namespace
ip netns exec netns1 ip link set veth1 netns 1
```


### Network Policy

**[Kubernetes 中的服务发现与负载均衡](https://mp.weixin.qq.com/s/klc0GYAcTthPdUaF-O7izQ)**:


**[Kubernetes 网络模型进阶](https://mp.weixin.qq.com/s/Jm8VynGd506wN5-yiLHzdg)**:


**[理解 CNI 和 CNI 插件](https://mp.weixin.qq.com/s/sGTEp9m8PC2zhlEgcnqtZA)**:


