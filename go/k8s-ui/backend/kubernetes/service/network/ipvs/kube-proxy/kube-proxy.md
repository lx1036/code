
**[ipvs in kube-proxy](https://github.com/kubernetes/kubernetes/blob/master/pkg/proxy/ipvs/proxier.go)**

**[Kube-Proxy IPVS 模式源码分析](https://xigang.github.io/2019/07/28/kube-proxy-source-code/)**



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
DaemonSet 形式跑在每一个 Node 节点上，当然也可以只跑在一小部分 Node 节点上转发数据包，只是代理节点，而另一大部分 Node 节点跑 kubelet 进程，作为计算节点：

![kube-proxy-size](./imgs/kube-proxy-size.png)

## 代码流程

* 实例化一个 kube-proxy command: **[app.NewProxyCommand()](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/proxy.go#L37)** ，然后执行 **[Options](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L104-L139)** 对象的 **[Run() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L496)** ，
Run() 函数内实例化一个 **[ProxyServer 对象](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L314-L317)**，然后调用 runLoop() 函数，该函数实际上主要是在 **[goroutine 内调用 ProxyServer 对象的 Run() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server.go#L334-L338)**，然后进入阻塞，直到 channel 来告诉进程停止工作。

* **[在实例化 ProxyServer 对象过程中](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server_others.go#L75-L382)** ，这300行左右代码中，重点是根据 `proxyMode` 变量选择是 iptables 还是 ipvs，这里主要看 ipvs 模式，而且只关注 ipv4 不考虑 ipv6，
这里重点是 **[proxier 对象实例化](https://github.com/kubernetes/kubernetes/blob/v1.18.2/cmd/kube-proxy/app/server_others.go#L307-L336)** ，它会调用 ipvs 包的 **[实例化逻辑](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L319-L482)** ，
进而后面会周期性的 **[调用该 proxier 对象的 syncProxyRules() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L479)** ，syncProxyRules() 函数可以说是整个 kube-proxy 的最核心的逻辑。

* **[syncProxyRules() 函数](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L989-L1626)** 这六百多行代码是整个 kube-proxy 模块的最核心的逻辑，会把用户创建的 service 转换为 ipvs rules，然后调用 **[ipvs go 客户端](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/util/ipvs/ipvs.go)** 写入内核中。
这里会根据每一个 service 去构建 **[ipvs rules](https://github.com/kubernetes/kubernetes/blob/v1.18.2/pkg/proxy/ipvs/proxier.go#L1115-L1540)** 。

以上是kube-proxy模块的大概流程，具体细节太复杂，不是一时半会就能看懂的！！！
