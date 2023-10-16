


# eBPF
安装辅助工具:
```shell
yum update -y && yum update -y iproute2 net-tools centos-release-scl llvm-toolset-7 bpftool && scl enable llvm-toolset-7 'bash' && clang --version && llc --version
```


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


## from-container bpf





# Cilium 基于 eBPF 收发包路径 datapath
http://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/ :
* L1->L2: packet 到达网卡，被网卡驱动轮询poll中执行
* L2->L3: poll->XDP, XDP 包括三种结果: pass/drop/transmit(XDP 使用 transmit 实现一个 TCP/IP 负载均衡器)，见 [L2-L3-XDP](./L2-L3-XDP-eBPF.png)，


> 如何查看已加载的 eBPF 程序，可参考 [Cilium Network Topology and Traffic Path on AWS](http://arthurchiao.art/blog/cilium-network-topology-on-aws/)

**[Life of a Packet in Cilium：实地探索 Pod-to-Service 转发路径及 BPF 处理逻辑](https://arthurchiao.art/blog/cilium-life-of-a-packet-pod-to-service-zh/)**

在 `/var/run/cilium/state/${CiliumEndpointID}/` 目录中包含 `lxc_config.h` 和 `bpf_lxc.o`二进制文件(源码是 `bpf_lxc.c`)，`bpf_lxc.c` 
包含了 `from-container` 和 `to-container` tc eBPF 程序:
```shell
yum update -y && yum update iproute2 net-tools -y

# (1)给宿主机默认网卡 eth0 挂载 tc eBPF 程序
tc filter show dev eth0 ingress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[from-netdev] direct-action not_in_hw tag 524a2ea93d920b5f
tc filter show dev eth0 egress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[to-netdev] direct-action not_in_hw tag a04f5eef06a7f555

# (2)给 cilium 网卡 cilium_host/cilium_net 挂载 tc eBPF 程序
# 只有 cilium_host 网卡挂载了 ebpf 程序，cilium_net 网卡没有 tc egress，cilium_host 和 cilium_net 是一对 veth pair 
# https://github.com/cilium/cilium/blob/v1.10.5/bpf/bpf_host.c
tc filter show dev cilium_host ingress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[to-host] direct-action not_in_hw tag 7afe1afd2f393b1b
tc filter show dev cilium_host egress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[from-host] direct-action not_in_hw tag 9b2b3e068f78309b
tc filter show dev cilium_net ingress # https://github.com/cilium/cilium/blob/v1.10.5/bpf/bpf_host.c
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host_cilium_net.o:[to-host] direct-action not_in_hw tag 7afe1afd2f393b1b

# (2)给容器网卡宿主机侧 lxcdf20e7cb3f7b 挂载 tc eBPF 程序，容器侧没挂载
# 查看host namespace这侧的 tc filter ingress挂载的ebpf程序 https://github.com/cilium/cilium/blob/v1.10.5/bpf/bpf_lxc.c
tc filter show dev lxcdf20e7cb3f7b ingress # 表示 packets 从 container 出去时，会运行该 eBPF 程序
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_lxc.o:[from-container] direct-action not_in_hw tag ddd77f293c0a5e6a
tc filter show dev lxc28c29fad770b egress # 表示 packets 进入 container 时，会运行该 eBPF 程序
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_lxc.o:[to-container] direct-action not_in_hw tag 610fb24c15536f9b
# 查看容器内eth0网卡有没有挂载 tc ingress/egress eBPF 程序，貌似没有挂载
docker inspect 8dba621397f0 | grep Pid
nsenter -t 18421 -n tc filter show dev eth0 ingress
nsenter -t 18421 -n tc filter show dev eth0 egress
nsenter -t 18421 -n arp -n # 查看 arp(Address Resolution Packet)
#Address                  HWtype  HWaddress           Flags Mask            Iface
#10.216.136.172           ether   92:36:a1:12:9b:1b   C                     eth0
#10.208.40.96             ether   92:36:a1:12:9b:1b   C                     eth0

yum install -y bpftool
# 查看所有loaded bpf程序
bpftool prog
bpftool prog dump xlated id 4409 # dump the interpreted BPF code
bpftool prog dump jited id 4409 # dump the JITed BPF code
```

使用 LLVM 编译 bpf 程序:
```shell
# To enable eBPF/eBPF JIT support
echo 1 > /proc/sys/net/core/bpf_jit_enable
clang -O2 -emit-llvm -c bpf.c -o - | llc -march=bpf -filetype=obj -o bpf.o

```


## BPF Map Types
* Hash tables
* Arrays
* LRU(Least recently used)
* Ring Buffer
* Stack trace
* LPM(Longest prefix match)

BPF Helpers:
* bpf_get_prandom_u32()
* bpf_skb_store_bytes()
* bpf_redirect()
* bpf_get_current_pid_tgid()
* bpf_perf_event_output()

eBPF Tail Calls 作用:
* chain programs together
* split programs into independent logical components
* make bpf programs compasable

eBPF Function Calls 作用：
* reuse inside of a programs
* reduce programs size


## 常用 BPF 工具

### bpftool
bpftool 是 linux 内核自带的工具，用来比如 attach BPF 程序到虚拟文件系统

```shell
# https://manpages.ubuntu.com/manpages/focal/man8/bpftool-map.8.html
yum install -y bpftool

```

### BPF 虚拟文件系统
和 cgroup /sys/fs/cgroup 虚拟文件系统一样，把一些 BPF 对象通过虚拟文件系统表示出来，可见文档：**[Persistent BPF objects](https://lwn.net/Articles/664688/)**
或者查看 **[Object Pinning](http://arthurchiao.art/blog/cilium-bpf-xdp-reference-guide-zh/#14-object-pinning%E9%92%89%E4%BD%8F%E5%AF%B9%E8%B1%A1)** 


## 参考文献
**[深入理解 Cilium 的 eBPF 收发包路径(datapath)](http://arthurchiao.art/blog/understanding-ebpf-datapath-in-cilium-zh/)**

**[ebpf datapath in cilium](https://docs.cilium.io/en/v1.10/concepts/ebpf/intro/)**

**[Cilium：BPF 和 XDP 参考指南](http://arthurchiao.art/blog/cilium-bpf-xdp-reference-guide-zh/)**

**[cilium && cgroup ebpf](https://speakerdeck.com/rueian/cilium-and-cgroup-ebpf)**

## 常见缩写
ELF: executable and linkable format
