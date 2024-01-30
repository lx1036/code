



# vmlinux.h
生成 vmlinux.h 文件头

```shell
# ubuntu 22.04 里创建成功, 20.04 创建不成功
bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

```




# 参考文献

xdp-l4lb:
* https://github.com/MageekChiu/xdp4slb/tree/dev-0.2
* https://github.com/Netronome/bpf-samples/tree/master/l4lb
* https://www.ebpf.top/post/xdp_lb_demo/
* conntrack: https://github.com/cilium/cilium/blob/main/bpf/lib/conntrack.h
* conntrack: https://github.com/l3af-project/eBPF-Package-Repository/blob/main/connection-limit/README.md
* traffic mirror: https://lfnetworking.org/open-sourcing-traffic-mirroring-ebpf-package-to-the-l3af-project/
* traffic mirror: https://github.com/l3af-project/eBPF-Package-Repository/tree/main/traffic-mirroring

iptables nat:
* https://www.jianshu.com/p/f7e50352e4ec
* https://serverfault.com/questions/586486/how-to-do-the-port-forwarding-from-one-ip-to-another-ip-in-same-network



