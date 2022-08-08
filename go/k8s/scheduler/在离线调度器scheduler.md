
> 按照自己的理解，实现一个 pod scheduler!!!

# pod scheduler
scheduler 既可以调度在线 pod，也可以调度离线 pod(dynamic scheduler)。

## dynamic scheduler 离线调度器 
使用 dynamic scheduler 来调度离线 pod:
基于扩展资源colocation/cpu 和 colocation/memory 实现离线任务的动态调度，优先将离线任务调度到节点负载较低、离线任务较少的混部节点上，
均衡不同节点之间的负载、减少业务之间的资源竞争。




# 参考文献
**[k8s v1.24.3](https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/scheduler)**

**[一文洞悉kubernetes资源调度机制](https://zhuanlan.zhihu.com/p/541025604)**
