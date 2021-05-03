

# metrics-server
作用：主要用来获取 node/pod 的 cpu/memory resource usage data。可以供 HPA 等弹性伸缩组件作为基础数据。
原理：读取 kubelet summary api，按照一定格式输出 resource usage data。


## 基本概念
k8s metrics api types definitions/clients 定义代码：https://github.com/kubernetes/metrics, 目前支持到 v1beta1。

CPU单位：
cpu的单位有时是n (1m = 1000*1000n )，见代码：k8s.io/apimachinery/pkg/api/resource/suffix.go::L123-L125

```shell
kubectl proxy

# 查看 metrics-server 支持的 metrics 类型
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1
curl http://127.0.0.1:8001/apis/custom.metrics.k8s.io/v1beta1
curl http://127.0.0.1:8001/apis/external.metrics.k8s.io/v1beta1


# 查看具体 pod/node metrics 数据
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/nodes
# 等同于
kubectl get nodes.metrics.k8s.io {node_name} -o yaml
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/nodes/{node_name}
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/pods
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/namespaces/default/pods
# 等同于
kubectl get pods.metrics.k8s.io -n default
curl http://127.0.0.1:8001/apis/metrics.k8s.io/v1beta1/namespaces/default/pods/example-app-64547d7dc-ffv9k
```





## 参考文献
**[metrics-server](https://github.com/kubernetes-sigs/metrics-server)**

**[资源指标管道](https://kubernetes.io/zh/docs/tasks/debug-application-cluster/resource-metrics-pipeline/)**

**[通过聚合层扩展 Kubernetes API](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)**

**[细说k8s监控架构](https://zhuanlan.zhihu.com/p/79732351)**

**[Getting started with developing your own Custom Metrics API Server](https://github.com/kubernetes-sigs/custom-metrics-apiserver/blob/master/docs/getting-started.md)**
