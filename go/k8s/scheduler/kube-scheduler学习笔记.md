




# kube-scheduler
目前扩展scheduler主要使用scheduler-framework。

## 基本概念
腾讯扩展的 scheduler，基于 Node 真实负载进行预选和优选：https://cloud.tencent.com/document/product/457/50843



## 参考文献
**[Scheduling enhancements 文档](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/OWNERS)**

**[scheduler community 文档](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduler.md)**

**[scheduler community 文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/scheduling/OWNERS)**

**[scheduler官方插件](https://github.com/kubernetes-sigs/scheduler-plugins)**



# descheduler
腾讯扩展的 descheduler，基于 Node 真实负载进行重调度的插件：https://cloud.tencent.com/document/product/457/50921


## 解决的问题
Kubernetes 的资源编排调度使用的是静态调度，将 Pod Request Resource 与 Node Allocatable Resource 进行比较，来决定 Node 是否有足够资源容纳该 Pod。
静态调度带来的问题是，集群资源很快被业务容器分配完，但是集群的整体负载非常低，各个节点的负载也不均衡。
descheduler可以把 running pods in node 给 move 到其他 nodes 上去，这点很重要!!!
