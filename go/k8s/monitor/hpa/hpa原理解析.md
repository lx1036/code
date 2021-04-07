

# HPA(Horizontal Pod Autoscaler)
HPA 会从 metrics-server 中获取 node/pods 的 cpu/memory metrics值，来计算副本值replicas，可以参考
文档：https://kubernetes.io/zh/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-metrics-apis ，
以及代码：
https://github.com/kubernetes/kubernetes/blob/v1.19.7/cmd/kube-controller-manager/app/options/hpacontroller.go
https://github.com/kubernetes/kubernetes/blob/v1.19.7/cmd/kube-controller-manager/app/autoscaling.go




## HPA 设计文档
**[HPA V2设计文档](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/autoscaling/hpa-v2.md)**
**[HPA UCloud 使用文档](https://docs.ucloud.cn/uk8s/bestpractice/autoscaling/hpa)**


## 参考文献
**[Kubernetes HPA 使用详解](https://www.qikqiak.com/post/k8s-hpa-usage/)**

