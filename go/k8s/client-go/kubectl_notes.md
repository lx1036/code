
### 列出有污点的nodes列表
```shell script
kubectl get nodes -o=custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers | awk -F '\t' '{for (i=1;i<=NF;i++){if ($i ~/map/) {print $i}}}'| sort | uniq
```

