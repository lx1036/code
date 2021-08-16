
# Volume Manager
Volume Manager：管理卷的 Mount/Unmount 操作、卷设备的格式化以及挂载到一些公用目录上的操作。
它主要是用来做本节点 Volume 的 Attach/Detach/Mount/Unmount 操作。

TODO: Attach/Detach 操作，比较下 VolumeManager 和 AttachDetachController 这块的逻辑对比。默认是 AttachDetachController 来处理这块逻辑。

**[Kubernetes Volume System Redesign Proposal](https://github.com/kubernetes/kubernetes/issues/18333)**
**[Detailed Design for Volume Attach/Detach Controller](https://github.com/kubernetes/kubernetes/issues/20262)**
**[Detailed Design for Volume Mount/Unmount Redesign](https://github.com/kubernetes/kubernetes/issues/21931)**
