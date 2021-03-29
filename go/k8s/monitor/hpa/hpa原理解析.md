

# HPA(Horizontal Pod Autoscaler)
HPA 会从 metrics-server 中获取 node/pods 的 cpu/memory metrics值，来计算副本值replicas，可以参考
文档：https://kubernetes.io/zh/docs/tasks/run-application/horizontal-pod-autoscale/#support-for-metrics-apis ，
以及代码：
https://github.com/kubernetes/kubernetes/blob/v1.19.7/cmd/kube-controller-manager/app/options/hpacontroller.go
https://github.com/kubernetes/kubernetes/blob/v1.19.7/cmd/kube-controller-manager/app/autoscaling.go



