
# VPC CNI
vpc-cni 解决 vpc 网络下容器网络通信问题。

VPC Route 路由模式：基于 vpc 路由表构建的容器网络组件，在该模式下，容器网络和节点网络不在一个网段。
这个模式和现在裸金属方案很类似：每一个 node 分配一个 pod cidr, 且 pod cidr > node 的路由会通过 cloud-controller-manager 写到 vpc 
路由表内。这与现在裸金属通过 BGP 宣告给交换机写动态路由方式不同。每一个 node 上通过 veth pair 和 route 来打通 
container net namespace <-> host net namespace。这样每一个 pod ip 在 vpc 内就是可达的。
优点：
    (1)无需考虑 service ip 在容器内不通这些问题(容器内iptables/ipvs 规则都是空的)，因为 packet 会到达 host net namespace，
经过宿主机侧的 iptables/ipvs 规则。
    (2)并且 pod tc egress/ingress 实现也会比较简单，egress 在容器侧的 eth0 tc 实现就行；ingress 在对端宿主机侧的 veth-xxx 网卡 tc 实现就行，
不像 ipvlan 那样不好做。

缺点：对 vpc 网络基础设施要求高，公司私有云在 vpc 路由表这块支持不够好。而且难点是也需要开发部署 cloud-controller-manager(在kube-controller-manager里)。

VPC ENI 弹性网卡模式(独占/共享 ENI)：基于 vpc 弹性网卡构建的 k8s cni 容器网络组件，该网络模式下，容器和节点分布在同一个网段。




# ENI(Elastic Network Interface)
参考自：https://github.com/AliyunContainerService/terway


# Aliyun API
使用Terway网络插件: https://help.aliyun.com/document_detail/97467.html

