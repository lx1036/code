

# EtcdServer

## 线性一致性读 ReadIndex
读请求流程：https://time.geekbang.org/column/article/335932

串行读：
线性读：



## 写流程
写流程大概：当 client 发起一个更新 hello 为 world 请求后，若 Leader 收到写请求，它会将此请求持久化到 WAL 日志，并广播给各个节点，
若一半以上节点持久化成功，则该请求对应的日志条目被标识为已提交，EtcdServer 模块异步从 Raft 模块获取已提交的日志条目，应用到状态机 boltdb 中。


