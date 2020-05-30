
# Logging / Monitoring 5%

1. 列出指定pod的日志中状态为Error的行，并记录在指定的文件上
```shell script
kubectl logs <podname> | grep Error > /opt/KUCC000xxx/KUCC000xxx.txt
```
