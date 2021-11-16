

## Cilium 如何解决 Pod 级别的 tc ingress/egress 带宽控制的？
bandwidth CNI Plugin 已经可以根据 Pod annotation "kubernetes.io/ingress-bandwidth"/"kubernetes.io/egress-bandwidth" 来 tc 设置
Pod 虚拟网卡的 ingress/egress。

但是，Cilium 下发了 tc eBPF，具体内容还未了解！！！
