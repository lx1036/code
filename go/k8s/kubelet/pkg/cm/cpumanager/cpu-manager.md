



## static policy
https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/cpu-manager.md#policy-2-static-cpuset-control

CPU 管理器不支持运行时下线和上线 CPUs。此外，如果节点上的在线 CPUs 集合发生变化，则必须驱逐节点上的 Pod，
并通过删除 kubelet 根目录中的状态文件 cpu_manager_state 来手动重置 CPU 管理器:

```shell
cat /var/lib/kubelet/cpu_manager_state
# 源码中 pkg/kubelet/cm/cpumanager/state/checkpoint.go::CPUManagerCheckpoint{}
# {"policyName":"none","defaultCpuSet":"","checksum":3242152201}
# {"policyName":"static","defaultCpuSet":"0,2-12,14-23","entries":{"235148fe-393f-47a8-a17d-bd55bc1a836b":{"cgroup1-0":"1,13"}},"checksum":1552716370}
```

cpu-manager static 策略中，只有 Guaranteed pod 中，指定了整数型 CPU requests 的容器，才会被分配独占 CPU 资源。

cpus 分为几组：
shared cpus: besteffort, burstable 和 non-integral guaranteed pod会占用shared cpus，作为default cpu set，会持久化到 /var/lib/kubelet/cpu_manager_state
reserved cpus: kube-reserved + system-reserved cpus
assignable cpus: shared - reserved，可分配的 cpus
exclusive cpus: 独占核，只被 integral guaranteed pod 独占的cpus，数据也会持久化到 /var/lib/kubelet/cpu_manager_state




## 参考文献
**[cpu-manager 设计文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/cpu-manager.md)**

**[深入理解 Kubernetes CPU Mangager](https://cloud.tencent.com/developer/article/1402119)**

**[控制节点上的 CPU 管理策略](https://kubernetes.io/zh/docs/tasks/administer-cluster/cpu-management-policies/)**

**[kubernetes kubelet组件中cgroup的层层"戒备"](https://www.cnblogs.com/gaorong/p/11716907.html)**
