

### kubelet 会周期性(10s)向 apiserver 上报 node status?
kubelet 对象会周期性(10s) pkg/kubelet/kubelet.go#L1362 去上报 node status pkg/kubelet/kubelet_node_status.go::syncNodeStatus()


