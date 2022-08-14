


**k8s volume 管理主要由三个组件来实现：VolumeManager、AttachDetachController 和 PVController。**

**[AttachDetachController 设计提案](https://github.com/kubernetes/kubernetes/issues/20262)**

一个 node 不应该负责管理一个 volume 的 attach/detach 操作，应该用一个 AttachDetachController 来独立处理这个逻辑。

为了向后兼容，到 k8s 1.19 为止，kubelet 还保留自己去执行 volume attach/detach 逻辑，通过参数 controllerAttachDetachEnabled(--enable-controller-attach-detach) 开启，
默认是 false 关闭的。

