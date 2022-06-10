

# Node SNAT
使用场景：对于 overlay 网络的 k8s 集群，pod 需要访问该 overlay 网络之外的公网或者公司内其他内网，需要 podIP 被 snat 成 nodeIP，回包则
是被 reverse 成 nodeIP -> podIP. **尤其对 cluster cidr 是私有 ip 需要这么做**。
在 vpc-cni 中经常使用，尤其是 vpc cidr 不能访问公网的(因为 cluster cidr 是私有网段比如 192.168.0.0/16)，但是 node 可以访问公网(vpc 内 node
访问公网方式，一般公有云提供常见的几种：每台 node 绑定弹性 IP；NAT 网关，包括 vpc cidr snat 公网IP，公网IP dnat vpc 内资源；负载均衡，该 vip 都是公网可访问的)。


## aws vpc-cni
可以看 aws pod to external communications: **[Pod to external communications](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/cni-proposal.md#pod-to-external-communications)**


## calico cni ippool natOutgoing
**[Configure outgoing NAT](https://projectcalico.docs.tigera.io/networking/workloads-outside-cluster)**


## bridge cni plugin 和 point-to-point veth pair cni plugin
ipMasq: 见 bridge.go 里已经做了 ipMasq，可以参考。
bridge 和 ptp plugin 都有 ipMasq 配置参数，用来配置 podIP SNAT nodeIP.







# 实现
使用 iptables 实现 SNAT.


# 参考文献
