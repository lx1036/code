



New() -> EventHandler(PriorityQueue) -> Schedule(Algorithm Schedule)


## 解决问题
dynamic-scheduler 需要根据该节点Node的 colocation/cpu 和 colocation/memory 来进行预选和优选Node节点。其中，colocation/cpu 和 colocation/memory
表示该 Node 给离线 Pod 所动态分配的 cpu/memory 资源，值是根据当前 available=(allocatable - Node上在线pod实际资源使用总和) ，乘以一个 ratio 计算出来的，即
cpu_isolation = (allocatable - cpu_usage) * ratio 。


### 开发参考文献
**[CPU 利用率提升至 55%，网易轻舟基于 K8s 的业务混部署实践](https://zhuanlan.zhihu.com/p/231631519)**
可以参见腾讯开发的 dynamic scheduler，基于 Node 真实负载进行重调度的插件: **[DynamicScheduler](https://cloud.tencent.com/document/product/457/50921)**





## 参考文献
**[自定义 Kubernetes 调度器](https://www.qikqiak.com/post/custom-kube-scheduler/)**

