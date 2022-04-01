
# Cilium Kubernetes IPAM

cilium ipam:kubernetes 工作机制：**[kubernetes ipam](https://docs.cilium.io/en/stable/concepts/networking/ipam/kubernetes/)**
cilium agent 会等待并 retrieveNodeInformation(corev1.node)，读取 node.Spec.PodCIDRs，没有则读取 node.Spec.PodCIDR，没有则
读取 node.Annotations["io.cilium.network.ipv4-pod-cidr"]，可以见代码：
* https://github.com/cilium/cilium/blob/v1.12.0-rc0/pkg/k8s/init.go#L190-L260
* https://github.com/cilium/cilium/blob/v1.12.0-rc0/pkg/k8s/node.go#L52-L206


# K8s NodeIPAM Controller
**[KEP-2593: Enhanced NodeIPAM to support Discontiguous Cluster CIDR](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/2593-multiple-cluster-cidrs)** : 该提议基本实现了我们的所有需求，
包括多个 ippool, 每个 node 可以多个 cidr，blockSize 可以根据 node 而变化，以及 ippool.Cidr 可以不连续的，等等。

