





CPU 管理器不支持运行时下线和上线 CPUs。此外，如果节点上的在线 CPUs 集合发生变化，则必须驱逐节点上的 Pod，
并通过删除 kubelet 根目录中的状态文件 cpu_manager_state 来手动重置 CPU 管理器:

```shell
cat /var/lib/kubelet/cpu_manager_state
# 源码中 pkg/kubelet/cm/cpumanager/state/checkpoint.go::CPUManagerCheckpoint{}
# {"policyName":"none","defaultCpuSet":"","checksum":3242152201}
```

cpu-manager static 策略中，只有 Guaranteed pod 中，指定了整数型 CPU requests 的容器，才会被分配独占 CPU 资源。






## 参考文献
设计文档: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/cpu-manager.md

**[深入理解 Kubernetes CPU Mangager](https://cloud.tencent.com/developer/article/1402119)**

**[控制节点上的 CPU 管理策略](https://kubernetes.io/zh/docs/tasks/administer-cluster/cpu-management-policies/)**

**[kubernetes kubelet组件中cgroup的层层"戒备"](https://www.cnblogs.com/gaorong/p/11716907.html)**
