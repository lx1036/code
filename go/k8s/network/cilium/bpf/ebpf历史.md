

# eBPF 历史

**[Why is the kernel community replacing iptables with BPF?](https://cilium.io/blog/2018/04/17/why-is-the-kernel-community-replacing-iptables)** :

* (1)iptables 最初提出来，主要来解决以下三个问题。但是 iptables rules 是线性处理，且缺少增量更新，每次更新一个 rule 需要把整个 
  rule list restore 出来再去 save 回去，这种全量更新方式：
  * Protect local applications from receiving unwanted network traffic (INPUT chain)
  * Protect local applications sending undesired network traffic (OUTPUT chain)
  * Filter network traffic forwarded/routed by a Linux system

* (2)为了解决 rule 越来越多的问题，也提出了一个优化方案：ipset。把 ip:port 以 hash 方式存入某个 named ipset，然后 iptables 直接用这个 named ipset。
  这样，一条 iptable rule 可以代替多条，可以有效缓解 rule 数量爆炸问题。但是，这不是最根本解决方案。

* (3)ipvs 相比于 iptables，ipvs 使用 hash 存储包过滤规，可以有效解决 rule 数量爆炸问题。

* (4)ebpf 可以替换 iptables，解决 net packet filter 这些问题，而且不会有 rule 数量爆炸问题。并且，有时候 packet 不需要被 copy 到内核走 netfilter，性能高。
  linux kernel 社区里，目前正在提议 add bpfilter 来替换 netfilter，见 **[net: add bpfilter](https://lwn.net/Articles/747504/)** 。

