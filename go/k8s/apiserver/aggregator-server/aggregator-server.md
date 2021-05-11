
# aggregator server
代码在：staging/src/k8s.io/kube-aggregator



## 基本概念
k8s 通过 api-aggregator 来支持扩展 kube-apiserver。通过以下options来开启 api-aggregator:

```shell
--requestheader-client-ca-file=<path to aggregator CA cert>
--requestheader-allowed-names=front-proxy-client
--requestheader-extra-headers-prefix=X-Remote-Extra-
--requestheader-group-headers=X-Remote-Group
--requestheader-username-headers=X-Remote-User
--proxy-client-cert-file=<path to aggregator proxy cert>
--proxy-client-key-file=<path to aggregator proxy key>
```



## 参考文献
**[通过聚合层扩展 Kubernetes API](https://kubernetes.io/zh/docs/concepts/extend-kubernetes/api-extension/apiserver-aggregation/)**

**[安装一个扩展的 API server](https://kubernetes.io/zh/docs/tasks/extend-kubernetes/setup-extension-api-server/)**

**[配置聚合层](https://kubernetes.io/zh/docs/tasks/extend-kubernetes/configure-aggregation-layer/)**

**[Kubernetes API Server Aggregator Server 架构设计源码阅读](https://cloudnative.to/blog/kubernetes-apiserver-aggregator-server/)**
