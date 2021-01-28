



## CSI Provisioner

代码：https://github.com/kubernetes-csi/external-provisioner

The external-provisioner is a sidecar container that dynamically provisions volumes by calling ControllerCreateVolume and ControllerDeleteVolume functions of CSI drivers.
The external-provisioner is an external controller that monitors PersistentVolumeClaim objects created by user and creates/deletes volumes for them.

