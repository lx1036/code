

## nsenter 命令使用
```shell

# 根据 net namespace 查询 pod ip
docker inspect ${container_id} | grep Pid # 6208
nsenter --target=${pid} --net # 进入这个pid的net namespace
ip addr # 可以查看到 eth0 的地址，即 pod ip

# 或者
nsenter --net=/proc/${pid}/ns/net -F -- ip -o -4 addr show dev eth0 scope global
nsenter --target=${pid} --net -F -- ip -o -4 addr show dev eth0 scope global

```


