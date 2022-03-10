
### 列出有污点的nodes列表
```shell script
kubectl get nodes -o=custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers | awk -F '\t' '{for (i=1;i<=NF;i++){if ($i ~/map/) {print $i}}}'| sort | uniq
```


# Node
```shell script
# drain node
kubectl drain <node-name> --ignore-daemonsets
# add/remove label
kubectl label nodes <node-name> <label-key>=<label-value>
kubectl label nodes <node-name> <label-key>-
# add/remove taint
kubectl taint nodes <node-name> <label-key>=<label-value>:NoSchedule
kubectl taint nodes <node-name> <label-key>-
# uncordon
kubectl uncordon nodes/<node-name>
```

