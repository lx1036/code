


# eBPF
eBPF go 包：https://github.com/cilium/ebpf
examples: https://github.com/cilium/ebpf/blob/master/examples/README.md


大规模微服务利器：eBPF + Kubernetes（KubeCon, 2020）: http://arthurchiao.art/blog/ebpf-and-k8s-zh/ (非常好的文章)
这篇文章讲清楚了 Cilium eBPF 比 kube-proxy 好在哪里，网络路径在哪里减少的:
* (1)减少了 netfilter 路径，直接由 tc ingress 跳转到 tc egress
* (2)如果 packet dst 不是本机，直接通过 XDP 在网卡中转出去，都不进入本机内核协议栈(eBPF 程序可以注入到网卡驱动中)


深入理解 Cilium 的 eBPF 收发包路径(datapath): http://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/



