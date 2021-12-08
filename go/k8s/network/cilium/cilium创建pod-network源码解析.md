

# Cilium 创建 pod network 源码解析

## Overview
我们生产K8s使用容器网络插件 Cilium 来创建 Pod network，下发 eBPF 程序实现 service 负载均衡来替换 kube-proxy，并且使用 BGP 协议来宣告路由给交换机，使得 pod ip 在内网可达。

目前 BGP speaker 使用 bird 软件， 不过随着 Cilium 最近新版本已经集成 **[MetalLB 库](https://github.com/cilium/metallb)**，可以使用 MetalLB 自带的 BGP speaker 来宣告路由，
后续只需要部署 cilium-operator deployment 和 cilium-agent daemonset 两个组件，无需部署 bird daemonset 组件，运维成本更低，
而且可以通过 cilium metrics 来获取 BGP 相关可观测性数据，具体详情可见 **[#16525](https://github.com/cilium/cilium/pull/16525/)** ，
代码可见 **[pod_cidr.go](https://github.com/cilium/cilium/blob/master/pkg/bgp/speaker/pod_cidr.go)** 。

Cilium 作为容器网络插件遵循 CNI 标准，部署时主要部署两个组件：cilium-operator deployment 和 cilium-agent daemonset。

cilium-operator 组件会根据选择的 ipam mode 不同选择不同的 ipam 逻辑，一般默认选择 cluster-pool 模式，这样 cilium-operator 会给每一个 v1.Node 对象
创建对应的 CiliumNode 对象，且根据 cluster-pool-ipv4-cidr 和 cluster-pool-ipv4-mask-size 两个配置计算出每一个节点的 pod cidr subnet 值，并存储在
该 CiliumNode 对象 spec.ipam.podCIDRs 字段中。比如给集群设置两个集群网段 cluster-pool-ipv4-cidr 为 10.20.30.40/24 和 50.60.70.80/24，
cluster-pool-ipv4-mask-size 设置为 26，cilium-operator 组件可以支持一个网段消耗完了可以从下一个网段继续分配 pod cidr net，代码见 
**[clusterpool.go#L107-L141](https://github.com/cilium/cilium/blob/master/pkg/ipam/allocator/clusterpool/clusterpool.go#L107-L141)** 。
根据 cluster-pool ipam 逻辑会依次把 cluster-pool-ipv4-cidr 10.20.30.40/24 分成 2^(26-24)=4 个子网段 pod cidr subnet，如果 10.20.30.40/24 消耗完了，
则继续消费 50.60.70.80/24 网段，一个子网段的切分可见代码 **[cidr_set.go](https://github.com/cilium/cilium/blob/master/vendor/github.com/cilium/ipam/cidrset/cidr_set.go)** ，
这个逻辑也是复用的 k8s nodeipam 逻辑 **[cidr_set.go](https://github.com/kubernetes/kubernetes/blob/v1.22.3/pkg/controller/nodeipam/ipam/cidrset/cidr_set.go)** 。
总之，cilium-operator 组件会根据 v1.node 的添加和删除，来添加和删除 CiliumNode 对象，同时使用 IPAM 来管理每一个 node 的子网 allocate/release 操作。

cilium-agent 组件会从其对应的子网段出再去划分出每一个 pod ip，上文已经说过 cilium-agent 或者 bird 会把每一个子网段 pod cidr subnet 通过 BGP 协议宣告给交换机，本机 node ip 作为下一跳，
所以使得每一个 pod ip 内网可达。kubelet 在创建 pod sandbox container 时会调用 cilium cni 二进制文件来为当前 pod 创建 network，
同时 cilium cni 作为客户端会调用服务端 cilium-agent pod 来分配 pod ip。

那 cilium-agent 做了哪些工作呢？以及默认不像 calico 那样每一个 pod 的宿主机端网卡，都会创建对应的路由，cilium 如何下发哪些 eBPF 程序做到这一点的？这是
本文重点讨论的问题。

## Cilium 工作原理
Cilium CNI 在为 pod 创建网络资源的过程，粗略说起来不复杂，主要分为三步：
* 为 pod 创建网卡 veth pair，并配置 mac 以及容器侧的路由等等网络资源
* Cilium IPAM 从该节点的子网段 pod cidr 中分配出一个 ip，并配置到 pod 网卡
* 为该 pod 创建对应的 CiliumEndpoint/CiliumIdentity 对象，计算并下发 network policy 规则，以及下发 eBPF 程序到 pod 网卡上。

限于篇幅以及为了精简，本文暂不考虑 network policy 逻辑，以及只考虑 create pod 时 cilium 的处理逻辑，不考虑 delete pod 时的逻辑。 

### Cilium 创建网络资源
在创建 pod 时，kubelet 会调用 Cilium 二进制文件，该二进制文件路径在启动 kubelet 时通过参数 `--cni-bin-dir` 传进来的，一般默认为 `/opt/cni/bin/` ，
比如在宿主机该目录下存在 `/opt/cni/bin/cilium-cni` 二进制文件，kubelet 启动参数 `--cni-conf-dir` 包含 cni 配置文件路径，一般默认为 `/etc/cni/net.d/05-cilium.conf` ，如文件内容为：

```json
{
  "cniVersion": "0.3.1",
  "name": "cilium",
  "type": "cilium-cni",
  "enable-debug": false
}
```

kubelet 和 cni 插件交互的具体内容可以参见之前的文章 **[Kubernetes学习笔记之Calico CNI Plugin源码解析(一)](https://juejin.cn/post/6916326439851655176)** 。

该文件内容将会被 json 反序列化为 **[NetConf](https://github.com/cilium/cilium/blob/master/plugins/cilium-cni/types/types.go#L22-L33)** 对象，
cni 代码中 **[args.StdinData 参数](https://github.com/cilium/cilium/blob/master/plugins/cilium-cni/cilium-cni.go#L278)** 即为该文件内容。

Cilium 支持两种 datapath types: veth pair 和 ipvlan，我们生产使用 veth pair，这样 Cilium 会为每一个 Pod 创建一个 veth pair，一个网卡在 host 侧，对端网卡在 container 侧。

#### 创建 veth pair 和路由
Cilium 调用 **[connector.SetupVeth()](https://github.com/cilium/cilium/blob/master/plugins/cilium-cni/cilium-cni.go#L389)** 创建 veth pair:

```go
veth, peer, tmpIfName, err = connector.SetupVeth(ep.ContainerID, int(conf.DeviceMTU), ep)
```

并且 host 侧网卡命名一般是: lxc + sha256(containerID))，如 lxc123abc；container 侧网卡命名一般是：tmp + maxLen(endpointID, 5)，如 tmp123,
并且设置：
* 设置 `/proc/sys/net/ipv4/conf/<veth>/rp_filter = 0`
* 设置两个网卡的 MTU
* 记录两个网卡的 mac 以及 interface name 和 interface index，供第三步创建 Endpoint 对象下发 eBPF 程序使用

然后移动 tmp123 网卡到 container netns，并重命名网卡为 eth0 网卡：

```go
_, _, err = connector.SetupVethRemoteNs(netNs, tmpIfName, args.IfName)
```

配置了 veth pair 后，开始配置容器侧的路由：

```go
ipConfig, routes, err = prepareIP(ep.Addressing.IPV4, false, &state, int(conf.RouteMTU))
// ...
if err = netNs.Do(func(_ ns.NetNS) error {
  // ...
    macAddrStr, err = configureIface(ipam, args.IfName, &state)
    return err
})
// ...
func IPv4Routes(addr *models.NodeAddressing, linkMTU int) ([]route.Route, error) {
    ip := net.ParseIP(addr.IPV4.IP)
    if ip == nil {
        return []route.Route{}, fmt.Errorf("Invalid IP address: %s", addr.IPV4.IP)
    }
    return []route.Route{
        {
            Prefix: net.IPNet{
                IP:   ip,
                Mask: defaults.ContainerIPv4Mask,
            },
        },
        {
            Prefix:  defaults.IPv4DefaultRoute,
            Nexthop: &ip,
            MTU:     linkMTU,
        },
    }, nil
}
```

容器侧的路由可以通过如命令查看容器侧路由表，其中网关地址 100.216.152.93 为 cilium_host 网卡的地址:

```shell
docker inspect 9a0874d84b93 | grep -i pid # 9a0874d84b93 为 container id
nsenter -t 15707 -n route -n
#Kernel IP routing table
#Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
#0.0.0.0         100.216.152.93   0.0.0.0         UG    0      0        0 eth0
#100.216.152.93   0.0.0.0         255.255.255.255 UH    0      0        0 eth0
```

总之，Cilium 会为 pod 创建 veth pair 和配置容器侧路由，这里默认没有配置宿主机侧路由，也没必要配置，因为可以通过 cilium_host 网卡的 tc eBPF 程序直接跳转到
宿主机网卡，无需通过 Linux 路由来跳转。总的来说，逻辑比较简单。

下一步 cilium cni 二进制文件作为客户端，调用本机 cilium daemon 获取 pod ip。

### Cilium IPAM 分配 pod ip
cilium cni 二进制调用 cilium daemon 服务端获取 pod ip，并把该 pod ip 配置到 pod 网卡上:

```go
ipam, err = c.IPAMAllocate("", podName, true)
```

可以通过如下命令查看容器网卡 ip:

```shell
nsenter -t 15707 -n ip addr
#1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
#    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
#    inet 127.0.0.1/8 scope host lo
#       valid_lft forever preferred_lft forever
#15723: eth0@if15724: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default 
#    link/ether 82:4a:3e:ac:af:0d brd ff:ff:ff:ff:ff:ff link-netnsid 0
#    inet 100.216.152.92/32 scope global eth0
#       valid_lft forever preferred_lft forever
```

pod ip 分配通过调用 cilium daemon IPAM 模块来从本机 pod cidr 子网段随机分配出 pod ip：

```shell
// pkg/client/ipam.go

// IPAMAllocate allocates an IP address out of address family specific pool.
func (c *Client) IPAMAllocate(family, owner string, expiration bool) (*models.IPAMResponse, error) {
	// ...
	resp, err := c.Ipam.PostIpam(params)
	// ...
	return resp.Payload, nil
}
```

下一步代码调用限于篇幅只给出调用路径：
* 客户端 ipam client **[ipam_client.go](https://github.com/cilium/cilium/blob/master/api/v1/client/ipam/ipam_client.go#L76-L108)**
* 服务端 ipam server **[post_ipam.go](https://github.com/cilium/cilium/blob/master/api/v1/server/restapi/ipam/post_ipam.go#L35-L61)**
* 服务端 ipam server ipam 模块逻辑 **[daemon/cmd/ipam.go](https://github.com/cilium/cilium/blob/master/daemon/cmd/ipam.go#L39-L82)**

主要代码如下，所以主要逻辑还是通过本机 pod cidr 中分配出一个 pod ip：

```go
func (h *postIPAM) Handle(params ipamapi.PostIpamParams) middleware.Responder {
    // ... 
	ipv4Result, ipv6Result, err := h.daemon.ipam.AllocateNextWithExpiration(family, owner, expirationTimeout)
}

func (d *Daemon) startIPAM() {
    // ...
    d.ipam = ipam.NewIPAM(d.datapath.LocalNodeAddressing(), option.Config, d.nodeDiscovery, d.k8sWatcher, &d.mtuConfig)
}

// 默认选择的是 cluster-pool ipam
func NewIPAM(nodeAddressing datapath.NodeAddressing, c Configuration, owner Owner, k8sEventReg K8sEventRegister, mtuConfig MtuConfiguration) *IPAM {
    // ...
    switch c.IPAMMode() {
    case ipamOption.IPAMKubernetes, ipamOption.IPAMClusterPool:
        if c.IPv4Enabled() {
            ipam.IPv4Allocator = newHostScopeAllocator(nodeAddressing.IPv4().AllocationCIDR().IPNet)
        }
}
```

通过一系列函数调用，会调用 hostScopeAllocator.AllocateNext() 来获取 pod ip:

```go
import (
    "github.com/cilium/ipam/service/ipallocator"
)

func newHostScopeAllocator(n *net.IPNet) Allocator {
    cidrRange, err := ipallocator.NewCIDRRange(n)

    a := &hostScopeAllocator{
        allocCIDR: n,
        allocator: cidrRange,
    }

    return a
}

func (h *hostScopeAllocator) AllocateNext(owner string) (*AllocationResult, error) {
	ip, err := h.allocator.AllocateNext()

	return &AllocationResult{IP: ip}, nil
}
```

和上文说到 cilium 使用 k8s 源码中从 cluster cidr 划分多个 pod cidr 一样，cilium 也是复用了 k8s 源码中从 pod cidr 中划分出一个个 pod ip 的逻辑，
cilium 为了防止引入其他 k8s 依赖包，单独把 k8s 源码中 **[ip allocator 逻辑](https://github.com/kubernetes/kubernetes/blob/v1.22.3/pkg/registry/core/service/ipallocator/allocator.go)**
单独出来一个包 **[cilium/ipam](https://github.com/cilium/ipam/blob/master/service/ipallocator/allocator.go)** 。

最后会调用 **[AllocationBitmap.AllocateNext()](https://github.com/cilium/ipam/blob/master/service/allocator/bitmap.go#L101-L114)** 从 pod cidr 中随机
分配出一个 pod ip:

```go
// AllocateNext reserves one of the items from the pool.
// (0, false, nil) may be returned if there are no items left.
func (r *AllocationBitmap) AllocateNext() (int, bool, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	next, ok := r.strategy.AllocateBit(r.allocated, r.max, r.count)
	if !ok {
		return 0, false, nil
	}
	r.count++
	r.allocated = r.allocated.SetBit(r.allocated, next, 1)
	return next, true, nil
}
```

该包支持 pod ip 两种分配策略：随机顺序分配 randomScanStrategy 和连续顺序分配 contiguousScanStrategy，默认使用随机顺序分配。

总之，cilium cni 二进制作为客户端调用 cilium daemon 服务端 IPAM 模块，来从 pod cidr 中随机分配获取 pod ip，并在上文第一步中配置 pod 网卡。总的来说，逻辑比较简单。

### Cilium 下发 eBPF 程序
最后 Cilium 针对每一个 pod 创建对应的 CiliumEndpoint 对象，在这一步会下发 tc eBPF 程序到 pod 网卡上:

```go
    if err = c.EndpointCreate(ep); err != nil {
		return
	}
```

下一步代码调用限于篇幅只给出调用路径：
* 客户端 endpoint client **[endpoint.go#L36-L42](https://github.com/cilium/cilium/blob/master/pkg/client/endpoint.go#L36-L42)**
* 客户端 endpoint client **[endpoint_client.go#L430-L465](https://github.com/cilium/cilium/blob/master/api/v1/client/endpoint/endpoint_client.go#L430-L465)**
* 服务端 endpoint server **[put_endpoint_id.go#L35-L64](https://github.com/cilium/cilium/blob/master/api/v1/server/restapi/endpoint/put_endpoint_id.go#L35-L64)**
* 服务端 endpoint server **[endpoint.go#L295-L551](https://github.com/cilium/cilium/blob/master/daemon/cmd/endpoint.go#L295-L551)**

本文跳过 network policy 创建过程，主要关注下发 eBPF 程序的逻辑 **[regenerate()](https://github.com/cilium/cilium/blob/master/pkg/endpoint/policy.go#L303-L419)**：

```go
func (d *Daemon) createEndpoint(ctx context.Context, owner regeneration.Owner, epTemplate *models.EndpointChangeRequest) (*endpoint.Endpoint, int, error) {
    // ...
	if build {
        ep.Regenerate(regenMetadata)
    }
    // ...
}

func (e *Endpoint) Regenerate(regenMetadata *regeneration.ExternalRegenerationMetadata) <-chan bool {
    epEvent := eventqueue.NewEvent(&EndpointRegenerationEvent{
        regenContext: regenContext,
        ep:           e,
    })
    
    resChan, err := e.eventQueue.Enqueue(epEvent)
    // ...
}

func (e *Endpoint) regenerate(context *regenerationContext) (retErr error) {
    revision, stateDirComplete, err = e.regenerateBPF(context)
    // ...
}
```

BPF 程序会被下发到宿主机 `/var/run/cilium/state` 目录中，regenerateBPF() 函数会重写 bpf headers，以及更新 BPF Map。更新 BPF Map 很重要，
下发到网卡中的 BPF 程序执行逻辑时会去查 BPF Map 数据:

```go
func (e *Endpoint) regenerateBPF(regenContext *regenerationContext) (revnum uint64, stateDirComplete bool, reterr error) {
	headerfileChanged, err = e.runPreCompilationSteps(regenContext)

	// 编译和加载 BPF 程序
    compilationExecuted, err = e.realizeBPFState(regenContext)
}

func (e *Endpoint) realizeBPFState(regenContext *regenerationContext) (compilationExecuted bool, err error) {
    // ...
    err = e.owner.Datapath().Loader().CompileAndLoad(datapathRegenCtxt.completionCtx, datapathRegenCtxt.epInfoCache, &stats.datapathRealization)
}
```

然后就是编译和加载 BPF 程序，Cilium 代码逻辑基本上就是执行类似如下命令：

```shell
# 编译 BPF C 程序
clang -O2 -emit-llvm -c bpf.c -o - | llc -march=bpf -filetype=obj -o bpf.o
# 下发 BPF 程序到容器网卡
tc filter add dev lxc09e1d2e egress bpf da obj bpf.o sec tc
```

Cilium 代码提供了 Loader 对象来编译和下发 BPF 程序，限于篇幅只给出调用路径：
* **[CompileAndLoad()](https://github.com/cilium/cilium/blob/master/pkg/datapath/loader/loader.go#L372-L402)**
* **[reloadDatapath()](https://github.com/cilium/cilium/blob/master/pkg/datapath/loader/loader.go#L282-L358)**
* **[reloadHostDatapath()](https://github.com/cilium/cilium/blob/master/pkg/datapath/loader/loader.go#L185-L274)** : 下发 tc eBPF 程序到 cilium_host/eth0 网卡
* **[replaceDatapath()](https://github.com/cilium/cilium/blob/master/pkg/datapath/loader/netlink.go#L58-L109)** : 下发 tc eBPF 程序到 pod 宿主机侧网卡

至此，以上代码逻辑已经编译并下发 BPF 程序到网卡。可以通过如下命令查看:

```shell
# 下发 bpf_lxc.c from-container 程序: https://github.com/cilium/cilium/blob/master/bpf/bpf_lxc.c#L970-L1025
tc filter show dev lxc3a01d529e083 ingress
#filter protocol all pref 1 bpf chain 0 
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_lxc.o:[from-container] direct-action not_in_hw tag b07a0188f79fbd7b

# 下发 bpf_host.c to-host 程序: https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L1106-L1188
tc filter show dev cilium_host ingress
#filter protocol all pref 1 bpf chain 0 
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[to-host] direct-action not_in_hw tag 7afe1afd2f393b1b

# 下发 bpf_host.c from-host 程序: https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L990-L1002
tc filter show dev cilium_host egress
#filter protocol all pref 1 bpf chain 0 
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_host.o:[from-host] direct-action not_in_hw tag 9b2b3e068f78309b

# 下发 bpf_host.c from-netdev 程序: https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L963-L988
tc filter show dev eth0 ingress
#filter protocol all pref 1 bpf chain 0 
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[from-netdev] direct-action not_in_hw tag 524a2ea93d920b5f

# 下发 bpf_host.c to-netdev 程序: https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L1004-L1104
tc filter show dev eth0 egress
#filter protocol all pref 1 bpf chain 0 
#filter protocol all pref 1 bpf chain 0 handle 0x1 bpf_netdev_eth0.o:[to-netdev] direct-action not_in_hw tag a04f5eef06a7f555 
```

假设容器中 `ping clusterip-service-ip`，出发走到另外一台机器的pod容器，会经过 from-container -> from-host -> to-netdev -> from-netdev -> to-host BPF 程序。

from-container BPF 可以参见 **[bpf_lxc.c#L970-L1025)](https://github.com/cilium/cilium/blob/master/bpf/bpf_lxc.c#L970-L1025)** ，主要处理来自容器方向的 packet， 
主要实现逻辑：
* validate_ethertype() 验证协议类型
* 如果是 ipv4，调用 tail_handle_ipv4()，进一步调用 handle_ipv4_from_lxc()，该函数主要完成：
  * 看看目标地址是否是 service ip，如果是则从 BPF Map 中找出一个 pod 作为目标地址，代码在 **[bpf_lxc.c#L559-L584](https://github.com/cilium/cilium/blob/master/bpf/bpf_lxc.c#L559-L584)** ，即实现了 service 负载均衡 
  * policy_can_egress4() 查看是否需要走 network policy，本文默认没有使用 network policy
  * ipv4_l3() 封包或者进行主机路由，设置 ttl 以及存储 src/dst mac 地址

```
static __always_inline int ipv4_l3(struct __ctx_buff *ctx, int l3_off,
				   const __u8 *smac, const __u8 *dmac,
				   struct iphdr *ip4)
{
	if (ipv4_dec_ttl(ctx, l3_off, ip4)) {
		/* FIXME: Send ICMP TTL */
		return DROP_INVALID;
	}

	if (smac && eth_store_saddr(ctx, smac, 0) < 0)
		return DROP_WRITE_ERROR;
	if (dmac && eth_store_daddr(ctx, dmac, 0) < 0)
		return DROP_WRITE_ERROR;

	return CTX_ACT_OK;
}
```

from-container 的经过 tc eBPF 后进入内核网络协议栈，上文介绍过容器内的路由网关是 cilium_host，packet 达到 cilium_host 网卡的 tc egress，
即走 from-host BPF 程序 **[bpf_host.c#L990-L1002](https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L990-L1002)** ，主要逻辑：
* 调用 from-netdev()
  * identity = resolve_srcid_ipv4() 解析这个包所属的 identity，从 ipcache map 中根据 IP 查询 identity
  * ctx_store_meta(ctx, CB_SRC_IDENTITY, identity) 把 identity 存储到 ctx->cb[CB_SRC_IDENTITY]。
  * ep_tail_call(ctx, CILIUM_CALL_IPV4_FROM_LXC) 尾调用 tail_handle_ipv4_from_netdev
  * handle_ipv4() 根据 dst_ip 查找 endpoint，即 pod ip

根据本机路由表，packet 会被转发给 eth0 网卡，会走 to-netdev BPF 程序 **[bpf_host.c#L1004-L1104)](https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L1004-L1104)** ，
该 BPF 程序只会处理 NodePort service 流量。本文暂不考虑 NodePort service packet。

经过以上 BPF 找到目标地址是另一台机器的 pod ip，包达到另一台机器，走 from-netdev BPF 程序，同样基本调用 from-netdev() 函数，逻辑和 from-host BPF 程序
基本一样，这里暂不赘述。

packet 到达 cilium_host 网卡走 to-host BPF 程序 **[bpf_host.c#L1106-L1188](https://github.com/cilium/cilium/blob/master/bpf/bpf_host.c#L1106-L1188)** ，
主要逻辑是把 packet 转发给其对应的 pod 网卡，这样无需走内核网络栈路由表了，性能更高：
* 调用 ctx_redirect_to_proxy_first()，然后调用 ctx_redirect_to_proxy_ingress4()，把 packet 转发给 pod 网卡，这样可以跳过内核协议栈 netfilter，性能更高


## 总结
通过本文可以知道，Cilium CNI 在创建 pod network 时主要做了三步：
* (1) 创建 pod 网络资源，包括 veth pair、路由以及配置 pod ip 等
* (2) cilium cni 调用 cilium daemon 从 pod cidr 中分配一个 pod ip，并配置到第一步中的 pod 网卡
* (3) 创建 CiliumEndpoint/CiliumIdentity 对象，计算 network policy，以及下发 BPF 程序到网卡。
Cilium 最重要的核心点就是 BPF 程序，包括实现了 service 负载均衡替换 kube-proxy、tc BPF ingress 跳转到 tc BPF egress 绕过 netfilter 实现高性能网络，等等功能。

总之，Cilium 主要使用了 BPF 技术实现了高性能网络，值得继续深入研究。
