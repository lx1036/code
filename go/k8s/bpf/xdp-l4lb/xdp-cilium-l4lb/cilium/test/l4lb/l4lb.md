


# L4LB DSR
(1)无法保留源 ClientIP 和 FullNAT；(2)保留源 ClientIP 和 DSR；
https://www.tigera.io/blog/introducing-the-calico-ebpf-dataplane/

DSR 缺点：
* 没有经过 LB gateway，这是不允许的，所以还是比较推荐 FullNat 模式，但是 FullNat 模式没法获取 ClientIP，TCP/UDP 只能通过 proxy-protocol 协议获取。
* 由于 Cilium 特地的 IP Option 可能会被底层网络拦截，所以 DSR 模式在一些公有云网络里无法使用。

