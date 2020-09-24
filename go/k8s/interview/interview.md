

# Kubernetes

## 第一阶段：基本知识
(1) 什么是 Kubernetes ？容器编排的价值和好处是什么？容器和主机部署应用的区别是什么？



## 第二阶段：技术细节
(1) Kubernetes 有哪些组件？

(2) 如何在 Kubernetes 中实现负载均衡？在生产中，你如何实现 Kubernetes 自动化？你如何扩展 Kubernetes 集群？


(3)你能解释 Deployment、ReplicaSets、StatefulSets、Pod、CronJob 的不同用途吗？
Kubernetes 如何处理持久性？服务和 ingress 的作用是什么？你何时会使用像 ConfigMap 或 secret 这样的东西？
Pod 亲和性作用是什么？你能举例说明何时使用 Init Container 么？
什么是 sidecar 容器？你能给出一个用例，说明你为什么要使用它么？


## 第三阶段：生产经验
(1) 在构建和管理生产集群时遇到的主要问题是什么？



## 第四阶段：K8S 的使用目的
(1) 什么是 Kubernetes Operator？


(2) 你对Kube-proxy有什么了解？


(3)kube-apiserver和kube-scheduler的作用是什么？
kube -apiserver遵循横向扩展架构，是主节点控制面板的前端。这将公开Kubernetes主节点组件的所有API，并负责在Kubernetes节点和Kubernetes主组件之间建立通信。
kube-scheduler负责工作节点上工作负载的分配和管理。因此，它根据资源需求选择最合适的节点来运行未调度的pod，并跟踪资源利用率。它确保不在已满的节点上调度工作负载。

(4) 你能简要介绍一下Kubernetes控制管理器吗？
多个控制器进程在主节点上运行，但是一起编译为单个进程运行，即Kubernetes控制器管理器。
因此，Controller Manager是一个嵌入控制器并执行命名空间创建和垃圾收集的守护程序。它拥有责任并与API服务器通信以管理端点。
node-controller,replication-controller,endpoint-controller,service-account/token controller

(5) 什么是ETCD？
Etcd是用Go编程语言编写的，是一个分布式键值存储，用于协调分布式工作。因此，Etcd存储Kubernetes集群的配置数据，表示在任何给定时间点的集群状态。

你对Kubernetes的负载均衡器有什么了解？
负载均衡器是暴露服务的最常见和标准方式之一。根据工作环境使用两种类型的负载均衡器，即内部负载均衡器或外部负载均衡器。
内部负载均衡器自动平衡负载并使用所需配置分配容器，而外部负载均衡器将流量从外部负载引导至后端容器。

什么是Ingress网络，它是如何工作的？

