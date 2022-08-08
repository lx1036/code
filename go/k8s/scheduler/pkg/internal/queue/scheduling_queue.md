


# Schedule Queue 调度队列
总的来说，Schedule Queue 就是个优先级队列，包含 backoffQ 和 activeQ：
* backoffQ 排队顺序是 pod 创建时间, 见 NewPriorityQueue()
* activeQ 排队顺序是 pod priority, 见 NewPriorityQueue()








## 参考文献

**[设计文档 Scheduling queue in kube-scheduler](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-scheduling/scheduler_queues.md)**

**[深入分析Kubernetes Scheduler的优先级队列](https://cloud.tencent.com/developer/article/1121557)**
