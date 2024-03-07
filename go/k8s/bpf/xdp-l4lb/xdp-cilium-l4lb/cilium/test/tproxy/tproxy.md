
# TProxy
内核代码:
```md
/root/linux-5.10.142/net/netfilter/xt_TPROXY.c
/root/linux-5.10.142/net/netfilter/nft_tproxy.c
https://www.kernel.org/doc/html/v5.8/networking/tproxy.html
https://powerdns.org/tproxydoc/tproxy.md.html

```


## eBPF for TProxy
内核代码：
```md
https://lore.kernel.org/bpf/20200329225342.16317-1-joe@wand.net.nz/

bpf_sk_assign 合并内核 commits:
summary: https://lore.kernel.org/bpf/20200329225342.16317-1-joe@wand.net.nz/
0/5: https://lore.kernel.org/bpf/CAADnVQJ5nq-pJcH2z-ZddDUU13-eFH_7M0SdGsbjHy5bCw7aOg@mail.gmail.com/
1/5 Add bpf_sk_assign eBPF helper: https://lore.kernel.org/bpf/20200329225342.16317-2-joe@wand.net.nz/
2/5 net: Track socket refcounts in skb_steal_sock(): https://lore.kernel.org/bpf/20200329225342.16317-3-joe@wand.net.nz/
3/5 bpf: Don't refcount LISTEN sockets in sk_assign(): https://lore.kernel.org/bpf/20200329225342.16317-4-joe@wand.net.nz/
4/5 selftests: bpf: add test for sk_assign: https://lore.kernel.org/bpf/20200329225342.16317-5-joe@wand.net.nz/
5/5 selftests: bpf: Extend sk_assign tests for UDP: https://lore.kernel.org/bpf/20200329225342.16317-6-joe@wand.net.nz/
```



