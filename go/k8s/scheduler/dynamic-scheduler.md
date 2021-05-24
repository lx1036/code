

# dynamic scheduler
动态调度，使用 dynamic scheduler 来调度离线 pod:
基于扩展资源colocation/cpu 和 colocation/memory 实现离线任务的动态调度，优先将离线任务调度到节点负载较低、离线任务较少的混部节点上，
均衡不同节点之间的负载、减少业务之间的资源竞争。

