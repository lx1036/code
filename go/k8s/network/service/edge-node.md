
# 问题

(2)什么是边缘节点？如何部署？
**[edge node](https://jimmysong.io/kubernetes-handbook/practice/edge-node-configuration.html)**
边缘节点：选择集群内部万兆网卡Node节点作为边缘节点，用来向集群外暴露服务。集群外部的服务通过调用边缘节点来调用集群内部的服务。
边缘节点需要考虑的两个问题：
* 边缘节点高可用，不能有单点故障，否则k8s集群不可用。
* 一堆边缘节点机器，必须只能有唯一一个外网访问IP和端口。

可以使用 keeplived 来实现。



