




# kube-scheduler
目前扩展scheduler主要使用scheduler-framework。

## 基本概念
腾讯扩展的 scheduler，基于 Node 真实负载进行预选和优选：https://cloud.tencent.com/document/product/457/50843



## kube-scheduler 插件 hooks
// @see https://github.com/kubernetes-sigs/scheduler-plugins/blob/release-1.19/pkg/noderesources/README.md

QueueSort: ["PrioritySort"]

PreFilter: ["NodeResourcesFit", "NodePorts", "PodTopologySpread", "InterPodAffinity", "VolumeBinding"]

Filter: ["NodeUnschedulable", "NodeResourcesFit", "NodeName", "NodePorts", "NodeAffinity", "VolumeRestrictions",
"TaintToleration", "EBSLimits", "NodeVolumeLimits", "VolumeBinding", "PodTopologySpread", "InterPodAffinity"]

PostFilter: ["DefaultPreemption"]

PreScore: ["InterPodAffinity", "PodTopologySpread", "TaintToleration"]

Score: ["NodeResourcesBalancedAllocation", "ImageLocality", "InterPodAffinity", "NodeResourcesLeastAllocated",
"NodeAffinity", "NodePreferAvoidPods", "PodTopologySpread", "TaintToleration"]

Reserve: ["VolumeBinding"]

Permit(目前还没有对应的 plugin)

PreBind: ["VolumeBinding"]

Bind: ["DefaultBinder"]


## 参考文献
**[Scheduling enhancements 文档](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/OWNERS)**

**[scheduler community 文档 Understanding the Kubernetes Scheduler](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduler.md)**

**[scheduler community 文档 Kubernetes Scheduler 设计文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling)**

**[scheduler官方插件](https://github.com/kubernetes-sigs/scheduler-plugins)**

**[Scheduling Framework 设计文档](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/624-scheduling-framework/README.md)**



# 混部系统 scheduler 设计
**[CPU 利用率提升至 55%，网易轻舟基于 K8s 的业务混部署实践](https://zhuanlan.zhihu.com/p/231631519)**

(1) 动态调度 dynamic scheduler：根据节点Node的真实负载实现离线业务的动态调度。这里是 cpu_isolation = (allocatable - cpu_usage) * ratio，这个 cpu_isolation 值
也是 isolation agent 执行绑核更新 cpuset 的依据。

可以参见腾讯开发的 dynamic scheduler，基于 Node 真实负载进行重调度的插件: **[DynamicScheduler](https://cloud.tencent.com/document/product/457/50921)**
组件原理：
3.5+1.8+0.6=5.9
组件包括：node-annotator 和 dynamic-scheduler


(2) 动态资源分配和隔离：根据在线业务的负载，动态调整分配给离线业务的资源量，动态执行资源隔离策略，降低甚至消除彼此之间的性能干扰。

