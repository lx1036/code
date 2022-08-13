

# VolumeBinding

StorageClass 支持 volumeBindingMode 包含两种模式：Immediate 和 WaitForFirstConsumer。
* Immediate: 表示 pvc 创建后，会立即由 PVController 自动去创建 pv，然后 bind pvc 和 pv。
* WaitForFirstConsumer: 表示 pvc 创建后，并不会立即由 PVController 自动去创建 pv，然后 bind pvc 和 pv，而是 PVController 去更新 pvc status 为 Pending。
只有在使用该 pvc 的 pod 被 Scheduler 调度后，才会去创建 pv 并 bind。主要就是由 Scheduler VolumeBinding plugin 做的。

这里主要研究 WaitForFirstConsumer volumeBindingMode 模式，主要链条为: VolumeBinding plugin 会在 PreBind 阶段给 pvc 加一个 annotation volume.kubernetes.io/selected-node: node1，
然后 PVController 根据这个 annotation 然后再给 pvc 加上 annotation volume.beta.kubernetes.io/storage-provisioner: csi.fusefs.com，
然后 external-provisioner sidecar 容器根据这个 annotation 再去创建对应的 pv，然后 PVController 周期性去 sync unbound pvc 里去做 bind pvc 操作。
而 VolumeBinding plugin 会每秒周期检查等待 bind pvc，发现 bind pvc 结束后则结束 PreBind 阶段。



