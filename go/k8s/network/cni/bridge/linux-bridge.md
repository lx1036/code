
# Linux bridge

## 如何连通两个 ns 之间的容器？

```shell
ip link add br0 type bridge
ip link set dev br0 up

ip netns add net3
ip link add br-veth0 type veth peer name br-veth1
ip link set dev br-veth1 netns net3
ip netns exec net3 ip link set dev br-veth1 name eth0
ip netns exec net3 ip addr add 20.0.3.10/24 dev eth0
ip netns exec net3 ip link set dev eth0 up
ip netns exec net3 ip route add default dev eth0

ip link set dev br-veth0 master br0
ip link set dev br-veth0 up
# 给 br0 添加 ip 20.0.3.11/24，会自动有路由 `20.0.3.0/24 dev br0 proto kernel scope link src 20.0.3.11`
ip addr add 20.0.3.11/24 dev br0

ip netns exec net3 ping -c 3 20.0.3.11
ping -c 3 20.0.3.10
```

Docker 容器网络就是 veth pair + bridge 模式组成的。
(1) 如何知道host上的 vethxxx 和哪个 container eth0是 veth pair 成对关系？
```shell script
# 在目标容器内
docker run -p 8088:80 -d  nginx
docker container ls
docker exec -it ${container_id} /bin/bash
cat /sys/class/net/eth0/iflink
# 在 host 宿主机上执行，两者结果应该是一样的，就表示 host 上的虚拟网卡 vethxxx 与容器内的 eth0 是一对
cat /sys/class/net/vethxxx/ifindex
```

(2) linux bridge
bridge 是一个虚拟网络设备，可以配置 IP、MAC 地址；其次，是一个虚拟交换机。
普通网络设备只有两个端口，如物理网卡，流量包从外部进来进入内核协议栈，或者从内核协议栈进来出去外面的物理网络中。
Linux bridge 则有多个端口，数据可以从任何端口进来，进来之后从哪个口出去取决于目的 MAC 地址，原理和物理交换机差不多。
```shell script
# 创建一个 bridge
# brctl addbr br1 # bridge-utils 软件包里的 brctl 工具管理网桥
# sudo apt install -y bridge-utils(Ubuntu)
# sudo yum install bridge-utils(CentOS)
ip link add name br0 type bridge
ip link list
ip link set br0 up
ip link list

# 创建一对 veth pair (eth0: 172.17.186.210)
ip link add br-veth0 type veth peer name br-veth1
ip addr add 172.17.186.101/24 dev br-veth0
ip addr add 172.17.186.102/24 dev br-veth1
ip link set br-veth0 up
ip link set br-veth1 up
# 把 br-veth0 搭到 br0 网桥上
ip link set dev br-veth0 master br0 # brctl addif br0 veth0
# 查看网桥上都有哪些网络设备
bridge link # brctl show
# 10: br-veth0 state UP @br-veth1: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 master br0 state forwarding priority 32 cost 2
# br-veth0 没法 ping 通 br-veth1
ping -c 1 -w 1 -I br-veth0 172.17.186.102
tcpdump -i lxc6e7eb5daff06 -nn tcp and port 80 # 抓包 tcp 且 80 端口

# veth0 的 IP 给 bridge
ip addr del 172.17.186.101/24 dev br-veth0
ip addr add 172.17.186.101/24 dev br0
```

**[Linux 虚拟网络设备详解之 Bridge 网桥](https://www.cnblogs.com/bakari/p/10529575.html)**


## bridge hairpin
bridge hairpin 定义：bridge 默认不允许包从收到包的端口发出，比如 bridge 从一个端口收到一个广播报文后，会将其广播到所有其他端口。而 bridge 
打开 haripin 模式后，允许从这个端口收到的包仍然从这个端口发出。这个特性用于NAT场景下，比如docker的nat网络，一个容器访问其自身映射到主机的端口时，
包到达bridge设备后走到ip协议栈，经过iptables规则的 dnat 转换后发现又需要从bridge的收包端口发出，需要开启端口的hairpin mode。
这个是 bridge 的一个概念。
```shell
brctl hairpin <bridge> <port> {on|off} turn hairpin on/off
```

### k8s service loopback
**hairpin 概念推广到 K8s 来说，说白了就是，一个 pod 访问一个 clusterIP/nodePort/externalIP service 后，经过 ipvs/iptables DNAT 后还是自己的 podIP，就是 hairpin
模式。** 所以，需要加一个 iptable snat rule 来把 src podIP 换成 src serviceIP。@see https://github.com/cloudnativelabs/kube-router/blob/master/pkg/controllers/proxy/network_services_controller.go#L1333-L1486 ，加了一个 iptable rule:
```shell
iptable -t nat -A ipvsHairpinChainName -s endpointIP/32 -d endpointIP/32 -m ipvs --vaddr serviceIP --vport servicePort -j SNAT --to-source serviceIP
```

