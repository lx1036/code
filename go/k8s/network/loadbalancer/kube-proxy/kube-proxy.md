
**[Kube-Proxy 的设计与实现](https://mp.weixin.qq.com/s/bYZJ1ipx7iBPw6JXiZ3Qug)**
**[Kube-Proxy Iptables 的设计与实现](https://mp.weixin.qq.com/s/oaW87xLnlUYYrwVjBnqeew)**
**[Kube-Proxy IPVS模式的原理与实现](https://mp.weixin.qq.com/s/RziLRPYqNoQEQuncm47rHg)**


## iptables
(1)snat:
```shell
iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -j SNAT --to-source 公网IP
```
这条命令的意思是将来自 10.10.0.0/16 网段的报文的源地址改为公司的公网 IP 地址。
* -t nat：表示 NAT 表
* -A POSTROUTING：表示将该条规则添加到 POSTROUTING 链的末尾，A 就是 append。
* -j SNAT：表示使用 SNAT 动作
* --to-source：表示将报文的源 IP 修改为哪个公网 IP 地址

(2)dnat
```shell
iptables -t nat -I PREROUTING -d 公网IP -p tcp --dport 公网端口 -j DNAT --to-destination 私网IP:端口号
```
这条命令的意思是将来自公网IP:端口号的报文的目的地址改为私网IP:端口，可以看到这里多了端口的信息。
原因是要区分公网访问的是私网的那个服务，所以需要明确到端口层级，才能精确送到客户端。而SNAT不需要端口信息也可以完成正确转发。
* -I PREROUTING：表示将该条规则插入到 PREROUTING 的首部，I 就是 insert
* --to-destination：表示将报文的目的 IP：端口修改为哪个私网IP：端口


# kube-proxy
概念：每台机器上都运行一个 kube-proxy 服务，它监听 API server 中 service 和 endpoint 的变化情况，并通过 iptables 等来为服务配置负载均衡（仅支持 TCP 和 UDP）。



# K8S 中的负载均衡
由kube-proxy实现的Service是一个四层的负载均衡，Ingress是一个七层的负载均衡。一个Service有对应的ClusterIP，即VIP，但在集群里不存在，没法ping通。





**[ipvs in kube-proxy](https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/ipvs/proxier.go)**

**[Kube-Proxy IPVS 模式源码分析](https://xigang.github.io/2019/07/28/kube-proxy-source-code/)**
**[浅谈 Kubernetes Service 负载均衡实现机制](https://www.infoq.cn/article/P0V9D4br7UDzWTgIHuYq)**


# (笔记)kube-proxy 源码中使用 ipvs 基本流程
学习 ipvs 过程中，大概翻了下 kube-proxy 的代码，简单笔记记录下。

> 源码主要在 proxy 包 https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/proxy.go 和
ipvs 包 https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go 两个文件夹内。

kube-proxy command 会调用 ipvs 包来写入 ipvs 的 virtual server 和 real server。

```shell script
# 在https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.18.md#downloads-for-v1182
# 下载 k8s v1.18.2 的 server binaries
wget https://dl.k8s.io/v1.18.2/kubernetes-server-linux-amd64.tar.gz
tar -zxf ./kubernetes-server-linux-amd64.tar.gz # 解压文件夹

# 下载 node binaries，也可以不下载
wget https://dl.k8s.io/v1.18.2/kubernetes-node-linux-amd64.tar.gz
tar -zxf ./kubernetes-node-linux-amd64.tar.gz
```

kube-proxy 二进制文件只有 37M 左右，不是很大，没记错的话 gitlab runner 有 57M 好像(跑 gitlab ci/cd，也是 go 写的)。kube-proxy 会以
DaemonSet 形式跑在每一个 Node 节点上，当然也可以只跑在一小部分 Node 节点上转发数据包，只是代理节点，而另一大部分 Node 节点跑 kubelet 进程，作为计算节点。

## 代码流程

* 实例化一个 kube-proxy command: **[app.NewProxyCommand()](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/proxy.go#L37)** ，然后执行 **[Options](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L104-L139)** 对象的 **[Run() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L496)** ，
  Run() 函数内实例化一个 **[ProxyServer 对象](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L314-L317)**，然后调用 runLoop() 函数，该函数实际上主要是在 **[goroutine 内调用 ProxyServer 对象的 Run() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L334-L338)**，然后进入阻塞，直到 channel 来告诉进程停止工作。

* **[在实例化 ProxyServer 对象过程中](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server_others.go#L75-L382)** ，这300行左右代码中，重点是根据 `proxyMode` 变量选择是 iptables 还是 ipvs，这里主要看 ipvs 模式，而且只关注 ipv4 不考虑 ipv6，
  这里重点是 **[proxier 对象实例化](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server_others.go#L307-L336)** ，它会调用 ipvs 包的 **[实例化逻辑](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L319-L482)** ，
  进而后面会周期性的 **[调用该 proxier 对象的 syncProxyRules() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L479)** ，syncProxyRules() 函数可以说是整个 kube-proxy 的最核心的逻辑。

* **[syncProxyRules() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L989-L1626)** 这六百多行代码是整个 kube-proxy 模块的最核心的逻辑，会把用户创建的 service 转换为 ipvs rules，然后调用 **[ipvs go 客户端](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/util/ipvs/ipvs.go)** 写入内核中。
  这里会根据每一个 service 去构建 **[ipvs rules](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L1115-L1540)** 。

以上是kube-proxy模块的大概流程，具体细节太复杂，不是一时半会就能看懂的！！！



# troubleshoot
(1) ipvs 模式为何必须把 vip 绑定到 kube-ipvs0 网卡上？
https://github.com/kubernetes/kubernetes/issues/107662#issuecomment-1017894646
https://blog.51cto.com/u_12790562/3799631
https://blog.csdn.net/yujia_666/article/details/111053291

ipvs DR 模式就必须要绑定 vip 到网卡上，这样 client 才可以外部 curl 一个 vip。而且：
> kube-ipvs0 should be down and noarp. This is intentional and not a bug. 
> kube-ipvs0 only purpose is to hold addresses that should be directed to ipvs. 
> Compare with proxy-mode=iptables where the vip addresses are on no interface at all. 
> Unfortunately there is no easy way to do that with ipvs, if it were, it would have been used and no kube-ipvs0 interface 
> would have been defined.

ipvs 是把 DNAT 放在 INPUT chain 上，而不是常用的 PREROUTING 上，所以必须把 vip 绑定在 kube-ipvs0 网卡上，让包进入 INPUT chain 上。
同时，kube-proxy 会增加一些 serviceIP -> kube-ipvs0 的路由。这样在 host 或在容器内 `curl -k https://192.168.0.1:443`，内核认为这个包访问自己，
会给 INPUT chain，ipvs 会在 INPUT chain 上做 DNAT serviceIP -> podIP，然后经过路由表判断从 eth0 出去。包会从 kube-ipvs0 网卡：

```shell
ip route list table local
local 192.168.0.1 dev kube-ipvs0 proto kernel scope host src 192.168.0.1
local 192.168.0.2 dev kube-ipvs0 proto kernel scope host src 192.168.0.2
local 192.168.0.2 dev nodelocaldns proto kernel scope host src 192.168.0.2
broadcast 192.168.0.2 dev nodelocaldns proto kernel scope link src 192.168.0.2
local 192.168.13.241 dev kube-ipvs0 proto kernel scope host src 192.168.13.241
local 192.168.17.214 dev kube-ipvs0 proto kernel scope host src 192.168.17.214
local 192.168.23.159 dev kube-ipvs0 proto kernel scope host src 192.168.23.159
local 192.168.62.122 dev kube-ipvs0 proto kernel scope host src 192.168.62.122
local 192.168.93.34 dev kube-ipvs0 proto kernel scope host src 192.168.93.34
local 192.168.115.35 dev kube-ipvs0 proto kernel scope host src 192.168.115.35
local 192.168.158.10 dev kube-ipvs0 proto kernel scope host src 192.168.158.10
local 192.168.212.236 dev kube-ipvs0 proto kernel scope host src 192.168.212.236
```

ipvs 不支持 SNAT，所以需要 iptables 来做 SNAT，主要就是 (通过 `iptables -t nat -S` 查看):

```shell
-N KUBE-MARK-MASQ
-A KUBE-SERVICES ! -s 20.225.0.0/16 -m comment --comment "Kubernetes service cluster ip + port for masquerade purpose" -m set --match-set KUBE-CLUSTER-IP dst,dst -j KUBE-MARK-MASQ
-A KUBE-LOAD-BALANCER -j KUBE-MARK-MASQ
-A KUBE-NODE-PORT -p tcp -m comment --comment "Kubernetes nodeport TCP port for masquerade purpose" -m set --match-set KUBE-NODE-PORT-TCP dst -j KUBE-MARK-MASQ
-A KUBE-MARK-MASQ -j MARK --set-xmark 0x4000/0x4000 # KUBE-MARK-MASQ 给包打个 mark


-A KUBE-POSTROUTING -m mark ! --mark 0x4000/0x4000 -j RETURN
-A KUBE-POSTROUTING -j MARK --set-xmark 0x4000/0x0
-A KUBE-POSTROUTING -m comment --comment "kubernetes service traffic requiring SNAT" -j MASQUERADE # POSTROUTING chain 上做 SNAT，DNAT 一般在 PREROUTING 上做
```

参考资料
**[kubernetes service 和 kube-proxy详解](https://plantegg.github.io/2020/09/22/kubernetes%20service%20%E5%92%8C%20kube-proxy%E8%AF%A6%E8%A7%A3/)**

