





### Stop and remove all containers
```shell
docker container stop $(docker container ls -aq) && docker container rm $(docker container ls -aq) && rm -rf /data/kubernetes
# 安装master时容易造成etcd安装错误，删除etcd旧数据
rm -rf /data/kubernetes/var/lib/etcd
rm -rf /data/kubernetes

# 查看 lldp
lldpcli show nei
```
