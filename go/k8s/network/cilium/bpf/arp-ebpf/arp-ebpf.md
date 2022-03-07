

# ARP in Cilium eBPF

调查 Cilium ARP eBPF 程序是如何拦截 ARP 请求，使得 host 侧的 lxc 网卡 mac 是 cilium_host 的 ip 的？
以及，为何 arp 相应不是 cilium_host 的响应？






# 参考文献
**[arp demo](https://github.com/P4-Research/ebpf-demos/blob/master/arp-resp/install_xdp.sh)**
**[cilium arp ebpf](https://huweicai.com/cilium-container-datapath/)**
**[cilium arp ebpf](https://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/#2-ebpf-%E6%98%AF%E4%BB%80%E4%B9%88)**
**[cilium arp ebpf](https://arthurchiao.art/blog/firewalling-with-bpf-xdp/#21-l2-example-drop-all-arp-packets)**

