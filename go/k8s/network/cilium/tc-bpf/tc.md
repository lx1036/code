

## Cilium 如何解决 Pod 级别的 tc ingress/egress 带宽控制的？
bandwidth CNI Plugin 已经可以根据 Pod annotation "kubernetes.io/ingress-bandwidth"/"kubernetes.io/egress-bandwidth" 来 tc 设置
Pod 虚拟网卡的 ingress/egress。

但是，Cilium 下发了 tc eBPF，具体内容还未了解！！！

## Cilium Bandwidth Manager
cilium 1.9 开始支持：https://cilium.io/blog/2020/11/10/cilium-19#bwmanager
Bandwidth Manager: https://docs.cilium.io/en/latest/gettingstarted/bandwidth-manager/
《Linux 高级路由与流量控制手册（2012）》第九章 用 tc qdisc 管理 Linux 网络带宽: http://arthurchiao.art/blog/lartc-qdisc-zh/

Cilium 使用 Bandwidth Manager 来管理 TC 流量控制，使用 EDT(Earliest Departure Time) 而不是 TBF(Token Bucket Filter)，
所以不需要社区的 bandwidth CNI Plugin。




# 参考文献
Cilium 下发 TC BPF 来实现 Pod 带宽限速：http://arthurchiao.art/blog/advanced-bpf-kernel-features-for-container-age-zh/


