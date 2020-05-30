

# Core Concepts 19%

1. 列出 minikube node 上所有正在运行的 pod?
列出 minikube node 上可分配资源，包括可用 CPU/Memory？
```shell script
kubectl get pods -A --field-selector="spec.nodeName=minikube,status.phase!=Succeeded,status.phase!=Failed"
```






