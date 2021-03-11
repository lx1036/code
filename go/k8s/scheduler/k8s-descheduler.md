

## 问题
Kubernetes 的资源编排调度使用的是静态调度，将 Pod Request Resource 与 Node Allocatable Resource 进行比较，
来决定 Node 是否有足够资源容纳该 Pod。静态调度带来的问题是，集群资源很快被业务容器分配完，但是集群的整体负载非常低，各个节点的负载也不均衡。











## Pod QoS(Quality of Service)

**[k8s 应用优先级，驱逐，波动，动态资源调整](https://my.oschina.net/u/4330952/blog/3371457)**

| request 是否配置 | limits 是否配置 | 两者关系            | QoS    | 说明              |
|    ---          | ---            | ---                | ---    | ---              |
| 是  | 是  |  requests = limits  | Guaranteed | 所有容器的cpu和memory必须配置相同的request和limit |
| 是  | 是  |  requests < limits  | Burstable | 只要有容器配置cpu/memory的request和limit就行 |
| 是  | 否  |                     | Burstable | 只要有容器配置cpu/memory的request就行 |
| 否  | 是  |   | Burstable/Guaranteed | 如果只是配置了limit，k8s会自动补充request，和limit值一样。所以，如果所有容器都配置了limit，则Guaranteed；只有部分容器配置limit，则是Burstable|
| 否 | 否 |  | Best Effort |  所有容器都没配置request/limit | 
