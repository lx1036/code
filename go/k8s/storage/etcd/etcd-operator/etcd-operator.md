

# Etcd 产品化

**[阿里巴巴云原生 etcd 服务集群管控优化实践](https://developer.aliyun.com/article/783544)** ：
* 把etcd cluster按照服务质量分成三个类型：BestEffort, Burstable 和 Guaranteed

Etcd产品化几个模块：
(1)etcd cluster 生命周期管理
* etcd 集群创建，销毁，停止，升级，故障恢复等。
* etcd 集群状态监控，包括集群健康状态、member 健康状态，访问量，存储数据量等。
* etcd 异常诊断、预案、黑盒探测，配置巡检等。

(2)etcd cluster 数据管理
* etcd 数据备份及恢复: 
  * snapshot 方式传统冷备份: backup/restore 实现冷备份
  * learner 热备份：使用 raft learner 特性实现实时热备份
* etcd 脏数据清理：根据指定 etcd key 前缀删除垃圾 kv 的能力，降低 etcd server 存储压力。
* etcd 热点数据识别: 按照 etcd key 前缀进行聚合分析热点 key 的能力，另外还可以分析不同 key 前缀的 db 存储使用量。
* etcd 数据迁移：
  * snapshot 方式: 先备份snapshot，再恢复进行迁移
  * raft learner 模式：我们使用 raft learner 特性可以快速从原集群分裂衍生出新的集群实现集群迁移。
* 数据水平拆分



## 参考文献
**[etcd-operator 快速入门完全教程](https://www.infoq.cn/article/ufq29mfxctyg4axibtge)**

**[etcd-operator](https://github.com/coreos/etcd-operator)**

**[腾讯云推出云原生etcd服务](https://segmentfault.com/a/1190000024483928)**

**[社区另一个版本 Etcd Operator](https://github.com/improbable-eng/etcd-cluster-operator)** 
**[视频版本](https://www.youtube.com/watch?v=nyUe-3zmHRc)**
