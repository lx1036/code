

# Cilium CNI 创建 pod network 源码解析
Cilium CNI 创建 pod network 具体流程原理：https://arthurchiao.art/blog/cilium-code-cni-create-network/


```

types.LoadNetConf(args.StdinData) -> connector.SetupVeth(ep.ContainerID, int(conf.DeviceMTU), ep)
-> netlink.LinkSetNsFd(*peer, int(netNs.Fd())) -> connector.SetupVethRemoteNs(netNs, tmpIfName, args.IfName)
-> c.IPAMAllocate("", podName, true) -> c.Ipam.PostIpam(params) 

-> ipam:*models.IPAMResponse

-> configureIface(ipam, args.IfName, &state) -> c.EndpointCreate(ep) -> c.Endpoint.PutEndpointID(params)

```


## Overview
我们生产使用容器网络插件 Cilium 来创建 Pod network，并且使用 BGP 协议来宣告路由给交换机，使得 pod ip 在内网可达。

目前 BGP speaker 使用 bird 软件， 不过随着 Cilium 最近新版本已经集成 **[MetalLB 库](https://github.com/cilium/metallb)**，可以使用 MetalLB 自带的 BGP speaker 来宣告路由，
后续只需要部署 cilium-operator deployment 和 cilium-agent daemonset 两个组件，无需部署 bird daemonset 组件，运维成本更低，
而且可以通过 cilium metrics 来获取 BGP 相关可观测性数据，具体详情可见 **[#16525](https://github.com/cilium/cilium/pull/16525/)** ，
代码可见 **[pod_cidr.go](https://github.com/cilium/cilium/blob/master/pkg/bgp/speaker/pod_cidr.go)** 。

Cilium 作为容器网络插件遵循 CNI 标准，部署时主要部署两个组件：cilium-operator deployment 和 cilium-agent daemonset。

cilium-operator 组件会根据选择的 ipam mode 不同选择不同的 ipam 逻辑，一般默认选择 cluster-pool 模式，这样 cilium-operator 会给每一个 v1.Node 对象
创建对应的 CiliumNode 对象，且根据 cluster-pool-ipv4-cidr 和 cluster-pool-ipv4-mask-size 两个配置计算出每一个节点的 pod cidr subnet 值，并存储在
该 CiliumNode 对象 spec.ipam.podCIDRs 字段中。比如给集群设置两个集群网段 cluster-pool-ipv4-cidr 为 10.20.30.40/24 和 50.60.70.80/24，
cluster-pool-ipv4-mask-size 设置为 26，cilium-operator 组件可以支持一个网段消耗完了可以从下一个网段继续分配 pod cidr net，代码见 
**[clusterpool.go#L107-L141](https://github.com/cilium/cilium/blob/master/pkg/ipam/allocator/clusterpool/clusterpool.go#L107-L141)** 。
根据 cluster-pool ipam 逻辑会依次把 cluster-pool-ipv4-cidr 10.20.30.40/24 分成 2^(26-24)=4 个子网段 pod cidr subnet，如果 10.20.30.40/24 消耗完了，
则继续消费 50.60.70.80/24 网段，一个子网段的切分可见代码 **[cidr_set.go](https://github.com/cilium/cilium/blob/master/vendor/github.com/cilium/ipam/cidrset/cidr_set.go)** ，
这个逻辑也是复用的 k8s nodeipam 逻辑 **[cidr_set.go](https://github.com/kubernetes/kubernetes/blob/v1.22.3/pkg/controller/nodeipam/ipam/cidrset/cidr_set.go)** 。
总之，cilium-operator 组件会根据 v1.node 的添加和删除，来添加和删除 CiliumNode 对象，同时使用 IPAM 来管理每一个 node 的子网 allocate/release 操作。

cilium-agent 组件会从其对应的子网段出再去划分出每一个 pod ip，上文已经说过 cilium-agent 或者 bird 会把每一个子网段 pod cidr subnet 通过 BGP 协议宣告给交换机，本机 node ip 作为下一跳，
所以使得每一个 pod ip 内网可达。kubelet 在创建 pod sandbox container 时会调用 cilium cni 二进制文件来为当前 pod 创建 network，
同时 cilium cni 作为客户端会调用服务端 cilium-agent pod 来分配 pod ip。

那 cilium-agent 做了哪些工作呢？以及默认不像 calico 那样每一个 pod 的宿主机端网卡，都会创建对应的路由，cilium 如何下发哪些 eBPF 程序做到这一点的？这是
本文重点讨论的问题。

## Cilium 工作原理
Cilium CNI 在为 pod 创建网络资源的过程，粗略说起来不复杂，主要分为三步：
* 为 pod 创建网卡 veth pair，并配置 mac 以及容器侧的路由，             这里默认没有配置宿主机侧的路由，是通过 cilium_host 网卡 tc eBPF 
* Cilium IPAM 从该节点的子网段 pod cidr 中分配出一个 ip，并配置到 pod 网卡
* 为该 pod 创建对应的 CiliumEndpoint/CiliumIdentity 对象，并下发 eBPF 程序到 pod 网卡上


### Cilium 创建网络资源



### Cilium IPAM 分配 pod ip




### Cilium 下发 eBPF 程序

































# 笔记
(1) 根据 containerID 获取 PID
```shell
docker inspect ${containerID} | jq '.[] | .State | .Pid' # 27219
docker inspect ${containerID} | grep "Pid" # 27219

# nsenter -t <pid> -n <command>
nsenter -t 27219 -n
ip addr # 获得 Cilium 创建的 veth peer 在 container side 一侧
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
749: eth0@if750: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
    link/ether 22:72:9c:50:45:e1 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.216.136.178/32 scope global eth0
       valid_lft forever preferred_lft forever
```


## 参考文献
**[Cilium Code Walk Through: CNI Create Network](https://arthurchiao.art/blog/cilium-code-cni-create-network/)**

**[Life of a Packet in Cilium：实地探索 Pod-to-Service 转发路径及 BPF 处理逻辑](https://arthurchiao.art/blog/cilium-life-of-a-packet-pod-to-service-zh/)**

**[Cilium Code Walk Through Series](http://arthurchiao.art/blog/cilium-code-series/)**

**[L4LB for Kubernetes: Theory and Practice with Cilium+BGP+ECMP](http://arthurchiao.art/blog/k8s-l4lb/)**
