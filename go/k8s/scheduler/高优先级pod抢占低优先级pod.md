


# Kubernetes学习笔记pod优先级抢占源码分析

抢占逻辑：
(1)判断该 pod 是否可以抢占，比如该 pod 抢占策略是不可抢占的，就不抢占
(2)找出所有可以被抢占的节点，但最多100台节点，只要改节点上有 Pod 优先级比当前 Pod 优先级低，该节点就是可以被抢占的节点。
当然，该 Pod 得经过针对这个 Node 的 Filter 走一遍，比如该 Node 资源还够不够。
(3)根据规则找出最优的可被抢占的节点，比如可以被驱逐的 victims 数量最少，优先级总和最小，node 上高优先级 Pod 数量最少等等
(4)驱逐 victim pod






## 参考文献
**[Pod 优先级与抢占](https://kubernetes.io/zh/docs/concepts/configuration/pod-priority-preemption/)**

**[kube-scheduler 优先级与抢占机制源码分析](https://www.bookstack.cn/read/source-code-reading-notes/kubernetes-kube_scheduler_preempt.md)**

**[Pod Preemption in Kubernetes](https://github.com/kubernetes/design-proposals-archive/blob/main/scheduling/pod-preemption.md)**

**[cross node preemption plugin](https://github.com/kubernetes-sigs/scheduler-plugins/blob/master/pkg/crossnodepreemption/README.md)**

**[Promote Pod Priority and Preemption to GA](https://github.com/kubernetes/enhancements/blob/master/keps/sig-scheduling/268-priority-preemption/README.md)**
