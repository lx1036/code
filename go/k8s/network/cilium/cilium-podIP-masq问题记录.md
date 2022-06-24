

# Cilium Masquerading podIP 问题记录

## 背景
在版本升级 cilium v1.8.1 到 v1.11.1 时，导致业务 pod 报错连接 mysql 授权错误，经过排查发现连接 mysql server 的 clientIP 是
业务 pod 所在的 nodeIP，而不是默认的 podIP，因为 mysql server 只授权了当前 K8s 集群的 pod cidr，所以报错授权问题。

矛盾点在于使用 cilium v1.8.1 时，出机器时的 IP 还是 podIP，但是 v1.11.1 却是 nodeIP，进一步发现 v1.8.2 版本也是 nodeIP。
K8s 网络这块采用的是 Cilium + BGP 模式，podIP 在公司内网可达，所以希望的也是业务 pod 从当前节点出去 IP 应该是 podIP 才对，cilium 估计是做了 SNAT，把
podIP SNAT 成 nodeIP。


## 原因
原因在于 cilium 默认会做 podIP masq，可以参考官网文档 v1.8 ：**[masquerading](https://docs.cilium.io/en/v1.8/concepts/networking/masquerading/)**

我们部署的 cilium 配置里也配置了 `masquerade: true`，实际上 cilium 会默认配置值为 `true` :

```yaml
masquerade: 'true'
enable-bpf-masquerade: 'true'
native-routing-cidr: 10.20.30.0/24
```

升级 cilium v1.11.1 时我们还是用的以上配置, cilium 新版本这个老配置 `masquerade: true` 已经废弃，改用 `enable-ipv4-masquerade: true`，
cilium 默认开启 podIP masquerade，见代码：https://github.com/cilium/cilium/blob/v1.11.1/daemon/cmd/daemon_main.go#L679-L680

所以升级 cilium v1.11.1 时需要改下配置就解决问题了：

```yaml
enable-ipv4-masquerade: 'false'
enable-bpf-masquerade: 'false'
ipv4-native-routing-cidr: 10.20.30.0/24 # 新版本废弃 native-routing-cidr 配置，使用该配置，默认也是使用集群 pod cidr，和配置值 cluster-pool-ipv4-cidr 相同
```


为何 cilium v1.8.1 没有报这个问题？
尽管 cilium v1.8.1 我们使用的配置是 `masquerade: true`，但是这个版本有个 bug，导致配置了也不起作用，podIP Masq 也不会走对应的 ebpf SNAT 规则，
在版本 v1.8.2 里修复了这个 bug，所以 cilium v1.8.2 之后默认都是开启 podIP Masq，尽管这个不是我们想要的。
bug 修复代码见：https://github.com/cilium/cilium/pull/12456


如果 pod 访问的目标 ip 在 ipv4-native-routing-cidr 网段内，也不会走 podIP Masq，ebpf c 代码里会判断如果在该网段内就跳过不走 masq 逻辑。
这样 pod 相互访问不会走 podIP Masq，只有访问集群外网络时才会这样。
ebpf c 代码跳过 ipv4-native-routing-cidr 网段逻辑见：
https://github.com/cilium/cilium/blob/v1.11.1/pkg/datapath/linux/config/config.go#L505-L518
https://github.com/cilium/cilium/blob/v1.11.1/bpf/lib/nodeport.h#L1160-L1170

然后，包从容器出来，会经过 node 上 eth0 网卡，该网卡上下发了 ebpf SNAT 逻辑，会把 podIP SNAT 成 nodeIP，SNAT 逻辑代码函数见：
https://github.com/cilium/cilium/blob/v1.11.1/bpf/lib/nat.h#L504-L570
https://github.com/cilium/cilium/blob/v1.11.1/bpf/lib/nat.h#L322-L378
这里难点主要是如何使用 ebpf 代码去做 SNAT, cilium 这块代码值得学习，这里也是 cilium 核心逻辑之一。

最后，cilium 会把该 ebpf c 程序下发到 eth0 网卡 egress 出口侧(默认是 eth0 网卡，可以在 cilium daemon 里配置出口网卡)，
可以在然后一台 K8s node 上执行以下命令看到 `to-netdev` ebpf 程序，这里 `to-netdev` 可以理解为这块 ebpf c 程序的名字:
```shell
tc filter show dev eth0 egress
#filter protocol all pref 1 bpf chain 0
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[to-netdev] direct-action not_in_hw tag aed7375159f1f3a4
```

to-netdev ebpf c 程序可见，这是包从 eth0 网卡 egress 侧出去时走的逻辑：https://github.com/cilium/cilium/blob/v1.11.1/bpf/bpf_host.c#L1003-L1103

同理，from-netdev ebpf c 程序是包进入 eth0 网卡 ingress 侧走的逻辑：https://github.com/cilium/cilium/blob/v1.11.1/bpf/bpf_host.c#L962-L987 , 
这里主要是防火墙或者 BPF NodePort 才有用，比如我们这里的 podIP Masq 时 BPF NodePort 是开启的。


### iptables snat masquerading
cilium 除了使用 ebpf 来实现 snat masq，也可以使用下发 iptables 规则来实现，可以见代码: 
https://github.com/cilium/cilium/blob/v1.11.1/pkg/datapath/iptables/iptables.go#L1097-L1137

可以修改 cilium 配置，然后使用命令 `iptables -t nat -S CILIUM_POST_nat` 查看:
```yaml
enable-ipv4-masquerade: 'true'
enable-bpf-masquerade: 'false'
ipv4-native-routing-cidr: 10.20.30.0/24 # 新版本废弃 native-routing-cidr 配置，使用该配置，默认也是使用集群 pod cidr，和配置值 cluster-pool-ipv4-cidr 相同
```


下发的 iptables 规则类似如下：
```shell
iptables -t nat -S POSTROUTING
#-P POSTROUTING ACCEPT
#-A POSTROUTING -m comment --comment "cilium-feeder: CILIUM_POST_nat" -j CILIUM_POST_nat
#-A POSTROUTING -m comment --comment "kubernetes postrouting rules" -j KUBE-POSTROUTING
#-A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE


iptables -t nat -S CILIUM_POST_nat
#-N CILIUM_POST_nat
#-A CILIUM_POST_nat -s 20.30.137.0/25 -m set --match-set cilium_node_set_v4 dst -m comment --comment "exclude traffic to cluster nodes from masquerade" -j ACCEPT
#-A CILIUM_POST_nat -s 20.30.137.0/25 ! -d 10.216.136.0/21 ! -o cilium_+ -m comment --comment "cilium masquerade non-cluster" -j MASQUERADE
#-A CILIUM_POST_nat -m mark --mark 0xa00/0xe00 -m comment --comment "exclude proxy return traffic from masquerade" -j ACCEPT
#-A CILIUM_POST_nat -s 127.0.0.1/32 -o cilium_host -m comment --comment "cilium host->cluster from 127.0.0.1 masquerade" -j SNAT --to-source 20.30.137.116
#-A CILIUM_POST_nat -o cilium_host -m mark --mark 0xf00/0xf00 -m conntrack --ctstate DNAT -m comment --comment "hairpin traffic that originated from a local pod" -j SNAT --to-source 20.30.137.116

```

包走 netfilter POSTROUTING chain 时会首先跳到 CILIUM_POST_nat chain 走完该 chain 的所有 rules，再跳转到 KUBE-POSTROUTING chain 的所有 rules。
CILIUM_POST_nat chain 包含的 rules 如上，podIP Masq 的 rule 主要是这条，通过 iptables 很简单就能实现 podIP SNAT 成 nodeIP：
```shell
-A CILIUM_POST_nat -s 20.30.137.0/25 ! -d 10.216.136.0/21 ! -o cilium_+ -m comment --comment "cilium masquerade non-cluster" -j MASQUERADE
```

当然，eBPF 因为会跳过 netfilter，包不必再去拷贝到内核里走 netfilter，性能相比 iptables 更高，所以还是使用 eBPF 来实现 podIP Masq，如果需要的话。
不过，eBPF 虽然性能高，但实现复杂。


## 总结
cilium 默认使用 podIP Masq，这样当 pod 不是访问其他 pod 时，会把 podIP SNAT 为 nodeIP，尤其在 podIP 是私网不可达且访问集群外部资源时有用。
但是，由于我们采用 cilium + BGP 模式，podIP 在公司内网可达，不需要这个功能，所以需要配置关闭。

另外，一个坑是我们配置一直没有关闭这个功能，所以配置一直都是错的，只是因为 cilium v1.8.1 自己的 bug，导致 podIP Masq 没有开启而已。

## 待调研
cilium podIP Masq ebpf 逻辑共用的 NodePort service 实现，可以调研下 cilium 如何实现 NodePort service？


## 参考文献
**[Masquerading](https://docs.cilium.io/en/v1.8/concepts/networking/masquerading/)**
https://github.com/cilium/cilium/pull/12456
**[datapath: Enable BPF MASQ for veth mode in IPv4](https://github.com/cilium/cilium/commit/0962c029849168da34d88f57dd7d0b73876d823b)**

