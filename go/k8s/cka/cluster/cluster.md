
# Cluster Maintenance 11%

1. 获取集群内指定 node(如名为 minikube 的 node) 的 --all-namespaces 的 events?
获取 default namespace 内所有 normal events?
获取 default namespace 内所有 warning events?
```shell script
kubectl get events --field-selector="involvedObject.name=minikube" -A
```

2. 列出k8s可用的节点，不包含不可调度的和 NoReachable 的节点，并把数字写入到文件里
