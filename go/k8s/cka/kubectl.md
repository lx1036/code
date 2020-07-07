
# Node
```shell script
# label
kubectl label nodes <node-name> <label-key>=<label-value>
# add/remove taint
kubectl taint nodes <node-name> <label-key>=<label-value>:NoSchedule
kubectl taint nodes <node-name> <label-key>=<label-value>:NoSchedule-
```

