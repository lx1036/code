

# K8S 事件机制

**[分析kubernetes中的事件机制](https://silenceper.com/blog/202003/kubernetes-event/)**
k8s 的 event 源码在 **[events](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/client-go/tools/record/event.go)** 中。
event 对象定义在 **[event obj](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/core/v1/types.go#L5150-L5214)**
