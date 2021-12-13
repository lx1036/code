

# Containerd
代码：https://github.com/containerd/containerd
arch: https://github.com/containerd/containerd/blob/main/design/architecture.md


## Storage
### Content


### Snapshot


### Diff


## Metadata
### Images


### Containers


### Tasks


### Events


# Containerd Plugins
containerd 架构由36个 plugins 实现:
* (1) ID/Type: content/io.containerd.content.v1, 创建 /var/lib/containerd/io.containerd.content.v1.content/ingrest 目录, 
* (2) aufs/io.containerd.snapshotter.v1, linux 不支持 aufs
* (3) devmapper/io.containerd.snapshotter.v1, 默认没有配置, https://github.com/containerd/containerd/blob/main/snapshots/devmapper/README.md
* (4) native/io.containerd.snapshotter.v1, 创建 /var/lib/containerd/io.containerd.snapshotter.v1.native/snapshots 目录
* (5) overlayfs/io.containerd.snapshotter.v1，
* (6) zfs/io.containerd.snapshotter.v1，linux 默认不支持 zfs
* (7) bolt/io.containerd.metadata.v1，写个 dbversion 到 boltdb /var/lib/containerd/io.containerd.metadata.v1.bolt/meta.db 文件中
* (8) walking/io.containerd.differ.v1
* (9) scheduler/io.containerd.gc.v1，
* (10) introspection-service/io.containerd.service.v1，
* (11) containers-service/io.containerd.service.v1，容器元数据metadata存在boltdb meta.db 文件中 https://github.com/containerd/containerd/blob/main/services/containers/local.go
* (12) content-service/io.containerd.service.v1，
* (13) diff-service/io.containerd.service.v1，
* (14) images-service/io.containerd.service.v1，
* (15) leases-service/io.containerd.service.v1，
* (16) namespaces-service/io.containerd.service.v1，
* (17) snapshots-service/io.containerd.service.v1，
* (18) linux/io.containerd.runtime.v1，创建 /var/lib/containerd/io.containerd.runtime.v1.linux 和 /run/containerd/io.containerd.runtime.v1.linux 目录，
* (19) task/io.containerd.runtime.v2，创建 /var/lib/containerd/io.containerd.runtime.v2.task 和 /run/containerd/io.containerd.runtime.v2.task 目录，
* (20) cgroups/io.containerd.monitor.v1，代码在 https://github.com/containerd/containerd/blob/main/metrics/cgroups/cgroups.go
* (21) tasks-service/io.containerd.service.v1，
* (22) restart/io.containerd.internal.v1，
* (23) containers/io.containerd.grpc.v1，




## 参考文献
本地开发调试 containerd: https://zhuanlan.zhihu.com/p/422522890

