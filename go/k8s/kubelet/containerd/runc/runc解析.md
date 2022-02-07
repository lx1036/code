




## runc 基本使用
使用说明：https://github.com/opencontainers/runc

```shell

mkdir -p mycontainer/rootfs && cd mycontainer
docker export $(docker create busybox) | tar -C rootfs -xvf - # 把 busybox:latest 镜像文件放入 rootfs 目录下
docker export $(docker create nginx:1.17.8) | tar -C rootfs -xvf -
runc spec

```
