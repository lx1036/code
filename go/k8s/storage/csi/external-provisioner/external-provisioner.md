



## CSI Provisioner

本仓库代码合并以下两个仓库：
https://github.com/kubernetes-csi/external-provisioner
https://github.com/kubernetes-sigs/sig-storage-lib-external-provisioner



The external-provisioner is a sidecar container that dynamically provisions volumes by calling ControllerCreateVolume and ControllerDeleteVolume functions of CSI drivers.
The external-provisioner is an external controller that monitors PersistentVolumeClaim objects created by user and creates/deletes volumes for them.


> 该csi provisioner可以参考cephfs provisioner实现： https://github.com/kubernetes-retired/external-storage/blob/master/ceph/cephfs/cephfs-provisioner.go
