
https://kubernetes.io/docs/concepts/storage/storage-classes/#provisioner :
每个StorageClass都必须有一个provisioner，如cephfs provisioner、ceph rbd provisioner等，
来决定创建pv时使用什么volume plugin。

写一个provisioner demo示例：
https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner/blob/master/examples/hostpath-provisioner/README.md

# cephfs provisioner

文档：https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/README.md
代码：https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/cephfs-provisioner.go

