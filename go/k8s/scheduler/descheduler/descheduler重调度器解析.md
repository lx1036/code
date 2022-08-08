

# Descheduler

## 解决的问题
Kubernetes 的资源编排调度使用的是静态调度，将 Pod Request Resource 与 Node Allocatable Resource 进行比较，来决定 Node 是否有足够资源容纳该 Pod。
静态调度带来的问题是，集群资源很快被业务容器分配完，但是集群的整体负载非常低，各个节点的负载也不均衡。
descheduler可以把 running pods in node 给 move 到其他 nodes 上去，这点很重要!!!

