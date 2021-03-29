




## 基本概念

```shell
kubectl proxy
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/nodes
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/pods
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/namespaces/cattle-prometheus/pods/exporter-node-cluster-monitoring-7zrjz
```





## 参考文献
**[metrics-server](https://github.com/kubernetes-sigs/metrics-server)**

**[资源指标管道](https://kubernetes.io/zh/docs/tasks/debug-application-cluster/resource-metrics-pipeline/)**

**[通过聚合层扩展 Kubernetes API](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)**

**[细说k8s监控架构](https://zhuanlan.zhihu.com/p/79732351)**

**[Getting started with developing your own Custom Metrics API Server](https://github.com/kubernetes-sigs/custom-metrics-apiserver/blob/master/docs/getting-started.md)**
