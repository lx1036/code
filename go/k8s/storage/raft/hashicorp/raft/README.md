
# Raft 实现三步走

(1) leader election
node 起始是 follower，运行 follower loop，heartbeatTimeout=1000ms, 每 heartbeatTimeout/10 leader 会发起心跳。
如果距离最新的 lastContact 超过 heartbeatTimeout，则 follower 成为 candidate 发起 leader election，
给每一个 follower transport RequestVote 获取 grant vote。达到 quorum size 之后，则成为 leader，运行 leader loop，
然后每 heartbeatTimeout/10 给每一个 follower peer 发起心跳来维持 leader 地位。

candidate 之后，获取 grant vote 必要条件：
* 如果已经有 leader 且 leader != candidate，则 reject vote；
* 如果 term 小于 follower term，则 reject term；
* 同 term 时只能给该 candidate 投票有且仅有一次，否则 reject term；
* 如果 follower lastLogTerm > candidate lastLogTerm，则 reject term；
* 如果 follower lastLogTerm == candidate lastLogTerm，但是 follower lastLogIndex == candidate lastLogIndex 则 reject vote；


(2) log replication



(3) safety


