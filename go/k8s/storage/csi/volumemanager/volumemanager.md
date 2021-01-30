


PV Controller： 负责 PV/PVC 绑定及周期管理，根据需求进行数据卷的 Provision/Delete 操作；
代码：https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/volume/persistentvolume/pv_controller.go

Attach/Detach Controller：负责数据卷的 Attach/Detach 操作，将设备挂接到目标节点；
代码: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/volume/attachdetach/attach_detach_controller.go

Volume Manager:Kubelet 中的组件，负责管理数据卷的 Mount/Umount 操作（也负责数据卷的 Attach/Detach 操作，需配置 kubelet 相关参数开启该特性）、卷设备的格式化等等；
代码：https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/volumemanager/volume_manager.go
