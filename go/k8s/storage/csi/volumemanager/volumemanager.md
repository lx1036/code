


PV Controller： 负责 PV/PVC 绑定及周期管理，根据需求进行数据卷的 Provision/Delete 操作；
代码：https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/volume/persistentvolume/pv_controller.go

Attach/Detach Controller：负责数据卷的 Attach/Detach 操作，将设备挂接到目标节点；
代码: https://github.com/kubernetes/kubernetes/blob/master/pkg/controller/volume/attachdetach/attach_detach_controller.go

Volume Manager:Kubelet 中的组件，负责管理数据卷的 Mount/Umount 操作（也负责数据卷的 Attach/Detach 操作，需配置 kubelet 相关参数开启该特性）、卷设备的格式化等等；
代码：https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/volumemanager/volume_manager.go

> Volume Manager 实际上是 Kubelet 中一部分，是 Kubelet 中众多 Manager 的一个。它主要是用来做本节点 Volume 的 Attach/Detach/Mount/Unmount 操作。
> 它和 AD Controller 一样包含有 desireStateofWorld 以及 actualStateofWorld，同时还有一个 volumePluginManager 对象，主要进行节点上插件的管理。在核心逻辑上和 AD Controller 也类似，通过 desiredStateOfWorldPopulator 进行数据的同步以及通过 Reconciler 进行接口的调用。


VolumeManager -> CSI volume plugin


## 问题

(1)VolumeManager 通过运行 async loops 来识别该node上的pod的哪些volumes，需要被 attach/detach 和 mount/unmount？
VolumeManager runs a set of asynchronous loops that figure out which volumes need to be attached/mounted/unmounted/detached 
based on the pods scheduled on this node and makes it so.


