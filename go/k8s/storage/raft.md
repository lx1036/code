

# Raft 设计
问题列表：

(1) 脑裂问题：什么是分布式系统中的闹裂？面对闹裂，我们的解决办法是什么？为什么多数派选举协议可以避免脑裂？

(2) leader 选举问题：为什么 raft 协议中只允许一个 leader? 怎么保证在一个任期内只有一个 leader 的？
集群中的节点怎么知道一个新的 leader 节点被选出了？如果选举失败了，会发生什么？如果两个节点都拿到了同样的票数，怎么选 leader？
如果老任期的 leader 不知道集群中新 leader 出现了怎么办？随机的选举超时时间作用，如果去选取它的值？
为什么不选择日志最长的服务器作为 leader？

(3) 日志复制问题：为什么 raft 需要使用日志？节点中的日志什么时候会出现不一致？Raft 怎么去保证最终日志会一致的？
在服务器突然崩溃的时候，会发生什么事情？如果 raft 服务奔溃后重启了，raft 会记住哪些东西？
哪些日志条目 raft 节点不能删除？raft 日志是无限制增长的吗？如果不是，那么大规模的日志是怎么存储的？
基于 raft 的服务崩溃重启后，是如何恢复的？

(4) 性能问题: 什么是 raft 系统中常见的性能瓶颈？



# 参考文献
**[raft 小论文](https://raft.github.io/raft.pdf)**

**[raft 大论文](https://github.com/ongardie/dissertation)**