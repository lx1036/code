

# containerd-shim-runc-v2 进程
containerd-shim-runc-v2 进程是干什么的？？？
每一个容器，都是 containerd-shim-runc-v2 进程启动的。

```shell
/usr/bin/containerd-shim-runc-v2 -namespace k8s.io -id ${pod_id} -address /run/containerd/containerd.sock
```
