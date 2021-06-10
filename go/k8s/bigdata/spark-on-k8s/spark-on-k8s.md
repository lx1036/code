
# Spark-on-K8s

Spark on K8s 方案中：
1. Spark-Submit通过调用K8s API在K8s集群中启动一个Spark Driver Pod;
2. Driver通过调用K8s API启动相应的Executor Pod, 组成一个Spark Application集群,并指派作业任务到这些Executor中执行;
3. 作业结束后,Executor Pod会被销毁, 而Driver Pod会持久化相关日志,并保持在'completed'状态,直到用户手清理或被K8s集群的垃圾回收机制回收.


Spark Operator包括如下几个组件:
1. SparkApplication控制器, 该控制器用于创建、更新、删除SparkApplication对象,同时控制器还会监控相应的事件,执行相应的动作;
2. Submission Runner, 负责调用spark-submit提交Spark作业, 作业提交的流程完全复用Spark on K8s的模式;
3. Spark Pod Monitor, 监控Spark作业相关Pod的状态,并同步到控制器中;
4. Mutating Admission Webhook: 可选模块,基于注解来实现Driver/Executor Pod的一些定制化需求;
5. SparkCtl: 用于和Spark Operator交互的命令行工具


Apache Spark作为通用分布式计算平台，K8s作为资源管理器平台。





## 参考文献

**[Spark Operator浅析](https://developer.aliyun.com/article/726791)**

**[spark-operator](https://github.com/GoogleCloudPlatform/spark-on-k8s-operator)**

**[Spark on Kubernetes 的现状与挑战](https://developer.aliyun.com/article/712297)**
