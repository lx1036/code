
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

