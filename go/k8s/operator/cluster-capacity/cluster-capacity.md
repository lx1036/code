

**[cluster capacity analysis framework](https://github.com/kubernetes-sigs/cluster-capacity)**

# cluster capacity
该工具可以实时查询集群可以部署的Pod数量，从而帮助集群管理者决定是否增加机器资源等。
实现原理为通过分析集群中的可用资源(包括CPU，Memory，IO等)，再根据用户输入Pod的请求资源大小，计算出各个节点还能部署的Pod数量。
目的是评估一个集群的剩余容量，包括每一个Node节点的剩余资源容量，容量指标包括cpu、内存和磁盘容量。
