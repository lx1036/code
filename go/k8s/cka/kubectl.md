
# Node
```shell script
# add/remove label
kubectl label nodes <node-name> <label-key>=<label-value>
kubectl label nodes <node-name> <label-key>-
# add/remove taint
kubectl taint nodes <node-name> <label-key>=<label-value>:NoSchedule
kubectl taint nodes <node-name> <label-key>-
```

