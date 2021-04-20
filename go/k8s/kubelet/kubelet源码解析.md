

# Kubelet
* Kubelet at a high level should deal with Volumes (Storage CSI), Networking (CNI), Resources (Container Manager), 
Runtime (CRI), Metrics (cAdvisor) APIs only.

## kubelet architecture(kubelet启动流程解析)

* kubelet在启动的时候，会先初始化Container Runtime启动时需要依赖的kubelet模块(Kubelet::initializeRuntimeDependentModules()):
  cadvisor,containerManager,evictionManager,containerLogManager,pluginManager,shutdownManager。
  其中cadvisor模块主要用于收集、聚合、处理和导出有关正在运行的容器的信息,同时也提供了Node MachineInfo Discover。
  cadvisor manager提供了MachineInfo()方法,可以获取node节点的机器信息，包括：CPU，Mem，操作系统等。
* 





**[Kubelet 源码剖析](https://www.infoq.cn/article/YHI2wUZWYmjmtCVNWVUc)**
**[Kubelet源码分析](https://xigang.github.io/2018/05/05/kubelet/)**
**[Kubelet 启动流程分析](https://mp.weixin.qq.com/s/hrE3onW_cbAQLz-UzeSyWA)**


## pod eviction
https://kubernetes.io/docs/tasks/administer-cluster/out-of-resource

**[深入k8s：资源控制Qos和eviction及其源码分析](https://www.cnblogs.com/luozhiyun/p/13583772.html)**

### Pod QoS(Quality of Service)
参考源码逻辑：https://github.com/kubernetes/kubernetes/blob/release-1.20/pkg/apis/core/v1/helper/qos/qos.go#L35-L102
QoS 设计文档：https://github.com/kubernetes/community/blob/master/contributors/design-proposals/node/resource-qos.md :
主要通过pod request/limit 来判定pod QoS: Guaranteed/Burstable/Best Effort


**[k8s 应用优先级，驱逐，波动，动态资源调整](https://my.oschina.net/u/4330952/blog/3371457)**

| request 是否配置 | limits 是否配置 | 两者关系            | QoS    | 说明              |
|    ---          | ---            | ---                | ---    | ---              |
| 是  | 是  |  requests = limits  | Guaranteed | 所有容器的cpu和memory必须配置相同的request和limit |
| 是  | 是  |  requests < limits  | Burstable | 只要有容器配置cpu/memory的request和limit就行 |
| 是  | 否  |                     | Burstable | 只要有容器配置cpu/memory的request就行 |
| 否  | 是  |   | Burstable/Guaranteed | 如果只是配置了limit，k8s会自动补充request，和limit值一样。所以，如果所有容器都配置了limit，则Guaranteed；只有部分容器配置limit，则是Burstable|
| 否 | 否 |  | Best Effort |  所有容器都没配置request/limit | 




### Eviction Signals
https://kubernetes.io/docs/tasks/administer-cluster/out-of-resource/#eviction-signals

| eviction signal | description  |
|    ---          | ---            |
| memory.available | memory.available := node.status.capacity[memory] - node.stats.memory.workingSet |
| nodefs.available | nodefs.available := node.stats.fs.available |
| nodefs.inodesFree | nodefs.inodesFree := node.stats.fs.inodesFree |
| imagefs.available | imagefs.available := node.stats.runtime.imagefs.available |
| imagefs.inodesFree | imagefs.inodesFree := node.stats.runtime.imagefs.inodesFree |
| allocatableMemory.available | allocatable - workingSet (of pods), in bytes |
| pid.available | pid.available := node.stats.rlimit.maxpid - node.stats.rlimit.curproc |

kubelet 需要设置 eviction.threshold，比如 memory.available<10% ，还需要设置间隔时间(eviction-monitoring-interval)
来比较 eviction.threshold，默认10s。

https://github.com/kubernetes/kubernetes/blob/release-1.20/pkg/kubelet/apis/config/v1beta1/defaults_linux.go :
hard-eviction: --eviction-hard=memory.available<100Mi,nodefs.available<10%,imagefs.available<15%



