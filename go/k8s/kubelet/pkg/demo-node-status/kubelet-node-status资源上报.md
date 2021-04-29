

# 资源压缩
直接压缩 Pod 的 request/limit 值。


# 资源超卖
通过拦截 kubelet 向 apiserver 上报 node status 时，修改 node.status.allocatable 值来实现资源超卖。

### kubelet 会周期性(10s)向 apiserver 上报 node status?
kubelet 对象会周期性(10s) pkg/kubelet/kubelet.go#L1362 去上报 node status pkg/kubelet/kubelet_node_status.go::syncNodeStatus()


