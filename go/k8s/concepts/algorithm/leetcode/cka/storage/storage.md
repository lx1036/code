# Storage 7%

1. 列出集群所有的pv，并以 name 字段排序/并以 capacity 字段排序（使用kubectl自带排序功能）
```shell script
kubectl get pv
kubectl get pv/{pv} -o json

kubectl get pv --sort-by=.metadata.name
kubectl get pv --sort-by=.spec.capacity.storage
```

2. 创建一个1G可读可写的PV，挂载在宿主机的"/mnt/data"目录
```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: task-pv-volume
  labels:
    type: local
spec:
  storageClassName: manual
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
```
