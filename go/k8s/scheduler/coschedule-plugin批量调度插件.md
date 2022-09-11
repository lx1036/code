

# pod 批量调度 plugin(gang)
重点是 Permit 做文章，在 Permit hook 只有 pods 超过了 PodGroup minNumber，send a signal to permit the waiting pods,
这样才会进入 bind()，@see https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/scheduler/schedule_one.go#L165-L199




# 参考文献

**[Coscheduling based on PodGroup CRD](https://github.com/kubernetes-sigs/scheduler-plugins/tree/master/kep/42-podgroup-coscheduling)**

