
https://kubernetes.io/docs/concepts/storage/storage-classes/#provisioner :
每个StorageClass都必须有一个provisioner，如cephfs provisioner、ceph rbd provisioner等，
来决定创建pv时使用什么volume plugin。

写一个provisioner demo示例：
https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/examples/hostpath-provisioner/README.md



## cephfs provisioner

文档：https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/README.md
代码：https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/cephfs-provisioner.go



# StorageClass
StorageClass 支持 volumeBindingMode 包含两种模式：Immediate 和 WaitForFirstConsumer。
* Immediate: 表示 pvc 创建后，会立即由 PVController 自动去创建 pv，然后 bind pvc 和 pv。
* WaitForFirstConsumer: 表示 pvc 创建后，并不会立即由 PVController 自动去创建 pv，然后 bind pvc 和 pv，而是 PVController 去更新 pvc status 为 Pending。
  只有在使用该 pvc 的 pod 被 Scheduler 调度后，才会由 PVController 创建 pv 并 bind。主要就是由 Scheduler VolumeBinding plugin 做的。






## 参考文献
https://kubernetes.io/docs/concepts/storage/storage-classes
