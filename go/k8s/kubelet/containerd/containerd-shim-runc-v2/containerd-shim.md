

# containerd-shim-runc-v2 进程
containerd-shim-runc-v2 进程是干什么的？？？
每一个容器，都是由 containerd-shim-runc-v2 进程启动的，容器里的进程都是它的子进程，这样 container lifecycle 独立于 containerd daemon 进程。
见 https://github.com/containerd/containerd/blob/main/design/lifecycle.md 。

```shell
/usr/bin/containerd-shim-runc-v2 -namespace k8s.io -id ${pod_id} -address /run/containerd/containerd.sock
```
