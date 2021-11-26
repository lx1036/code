


# eBPF
eBPF go 包：https://github.com/cilium/ebpf
examples: https://github.com/cilium/ebpf/blob/master/examples/README.md


大规模微服务利器：eBPF + Kubernetes（KubeCon, 2020）: http://arthurchiao.art/blog/ebpf-and-k8s-zh/ (非常好的文章)
这篇文章讲清楚了 Cilium eBPF 比 kube-proxy 好在哪里，网络路径在哪里减少的:
* (1)减少了 netfilter 路径，直接由 tc ingress 跳转到 tc egress
* (2)如果 packet dst 不是本机，直接通过 XDP 在网卡中转出去，都不进入本机内核协议栈(eBPF 程序可以注入到网卡驱动中)


https://docs.cilium.io/en/v1.10/concepts/ebpf/intro/ :
* XDP eBPF Hook: 这个 hook 点在网络流量进入机器最开始处，主要可以用来加载一些过滤eBPF程序，如drop有害流量、DDOS处理等。
* Traffic Control Ingress/Egress eBPF Hook: 在网卡 tc ingress hook 点下发eBPF程序
* Socket eBPF Hook: 



# Cilium 基于 eBPF 收发包路径 datapath
http://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/ :
* L1->L2: packet 到达网卡，被网卡驱动轮询poll中执行
* L2->L3: poll->XDP, XDP 包括三种结果: pass/drop/transmit(XDP 使用 transmit 实现一个 TCP/IP 负载均衡器)，见 [L2-L3-XDP](./L2-L3-XDP.png)，


> 如何查看已加载的 eBPF 程序，可参考 [Cilium Network Topology and Traffic Path on AWS](http://arthurchiao.art/blog/cilium-network-topology-on-aws/)
```shell
yum update -y && yum update iproute2 -y
tc filter show dev cilium_net ingress
# 查看host namespace这侧的 tc filter ingress挂载的ebpf程序
tc filter show dev lxcdf20e7cb3f7b ingress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_lxc.o:[from-container] direct-action not_in_hw tag ddd77f293c0a5e6a
yum install -y bpftool
# 查看所有loaded bpf程序
bpftool prog
bpftool prog dump xlated id 4409 # dump the interpreted BPF code
bpftool prog dump jited id 4409 # dump the JITed BPF code
```




## 参考文献
**[深入理解 Cilium 的 eBPF 收发包路径(datapath)](http://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/)**

**[ebpf datapath in cilium](https://docs.cilium.io/en/v1.10/concepts/ebpf/intro/)**

**[Cilium：BPF 和 XDP 参考指南](http://arthurchiao.art/blog/cilium-bpf-xdp-reference-guide-zh/)**

