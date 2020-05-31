
# Cluster Maintenance 11%

1. 获取集群内指定 node(如名为 minikube 的 node) 的 --all-namespaces 的 events?
获取 default namespace 内所有 normal events?
获取 default namespace 内所有 warning events?
```shell script
kubectl get events --field-selector="involvedObject.name=minikube" -A
```

2. 列出k8s可用的节点，不包含不可调度的和 NoReachable 的节点，并把数字写入到文件里?


3. 如何为指定 namespace 配置默认的 cpu/memory request and limit?
```shell script
kubectl create namespace limit-range
kubectl delete namespace limit-range

kubectl --kubeconfig ./kubeconfig.yml create namespace limit-range
kubectl --kubeconfig ./kubeconfig.yml delete namespace limit-range
```
```yaml
# 创建 LimitRange 对象来设置默认的 request/limit
apiVersion: v1
kind: LimitRange
metadata:
  name: mem-limit-range
spec:
  limits:
  - default:
      memory: 512Mi
    defaultRequest:
      memory: 256Mi
    type: Container
```

4. 如何获取一个 Node 的 memory/cpu request and limit?
