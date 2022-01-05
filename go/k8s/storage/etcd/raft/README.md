
# Raft
算法论文简略版本：https://raft.github.io/raft.pdf

raft 共识算法主要分为三个问题：
* leader election: 集群中需要选出有且只有一个 leader
* log replication: leader 会把 log entry 发给每一个 follower，并且 follower 会提交 log entry 到状态机(mvcc boltdb)，不断紧追 leader 进度
* safety: 一些限制条件，比如 leader/follower 各个节点状态机 applied 应用的 log entry，其任意位置肯定是相同的；已经提交的 log entry 在 leader 重新选举中，必须在新 leader 中，并且新 leader term 肯定是更高的。等等一些限制条件来保证安全性。

raft 模块包括：共识模块，log memory/wal storage 模块。状态机是 mvcc boltdb 模块。见 ![raft arch](./raft-arch.png)

## 成员变更(raft conf change)



## Trouble shoot
(1) raft协议，leader在commit了一条日志后，立刻挂了，那其他节点如何处理这条日志？
https://www.zhihu.com/question/357207584

