

# k8s monitor arch
两种 metrics: system metrics(core metrics / non-core metrics) 和 service metrics

**[core metrics source](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/instrumentation/monitoring_architecture.md#core-metrics-pipeline)** : 
* kubelet, 包含 node/pod/container usage
* resource estimator


# kube-state-metrics(K8S exporter)
是一个 k8s prometheus exporter，暴露k8s内置对象的一些 metrics，比如 pod_container_status_running 等等，是k8s cluster state 的一个 snapshot 而已。
个人感觉，没有 metrics-server 更实用，metrics-server metrics 暴露的是 pod/nodes 的 cpu/memory usage，这个更实用。

缺点：
(1) 每次抓取都要耗费 10-20s 这么久
(2) 占用内存很高



## 参考文献
**[kube-state-metrics](https://github.com/kubernetes/kube-state-metrics)**
