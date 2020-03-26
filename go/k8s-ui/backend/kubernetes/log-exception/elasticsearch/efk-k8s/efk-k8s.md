


# EFK-K8S
K8S 官方有个 EFK 插件说明 **[Elasticsearch Add-On](https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/fluentd-elasticsearch/README.md)**:
Elasticsearch 是查询和存储 log 的搜索引擎。
Fluentd 把来自于 K8S 的日志发送给 Elasticsearch。
Kibana 展示存储在 Elasticsearch 的日志。

## Elasticsearch
以 StatefulSet 资源对象部署。

## Fluentd
以 DaemonSet 资源对象部署，每一个 worker 节点部署一个 pod。读取来自于 kubelet/container runtime/containers 的日志，然后发送给Elasticsearch。




