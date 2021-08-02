


## Leader Election

### Raft算法详解–选主(Leader election)
概述
本文详细介绍了Raft算法的选主过程，包括：选主的流程，选主过程中各个节点的角色转换和消息发送机制。通过本文的学习，可以掌握Raft算法选主机制。

术语说明
Leader election
共识算法中的选举领导者的过程。

Follower
跟随者。处于该角色的节点不参与选举过程。

Candidate
选举者。处于该角色的节点参与选举过程，可以为某个Leader投票。

Leader
领导者。处于该角色的节点管理整个集群。

AppendEntries RPC
RPC消息，由领导者发送，用来追加log记录，该请求也用来作为心跳消息。

Request-Vote RPC
RFC消息，投票请求，由Candidate发送。

Raft算法的选主过程详解
Raft使用心跳机制来触发选主(Leader election)过程。当服务节点启动时，它们的角色是follower(跟随者)。服务节点将会一直保持follower的状态，
直到接收到来自某个Leader(领导者)或candidate(选举者)的RPC请求。为了保持Leader服务节点的权限，Leader会周期性的发送心跳消息(内容是AppendEntries RPC请求，
该请求不包含log内容)给所有的follower。如果一个follower在一个时间周期内没有收到任何通信消息，称为选举超时，那么它认为没有有效的Leader并开始选举选择新的Leader。

为了发起选举，Follower会递增当前时间周期，并转换到candidate状态。然后，它会开始为它自己投票，并且并行地向集群中其他的所有服务节点发送RequestVote RPC请求。
该candidate会继续保持在这个状态中，直到发生了以下三件事之一：

它赢得了选举
集群中另一个节点赢得了选举
该时间周期内没有任何节点赢得选举(超时)

这些结果将会在下面的章节中进行讨论。

如果一个Candidate在同一时间周期内从整个群集中的大多数服务节点收到投票，则它赢得本次选举。在给定的时间周期内每个服务节点最多为一名候选人投票，投票生效的方式是：
先来先得(注意：5.4增加了一个额外的投票限制)。这个主要的规则确保在一个特定的时间周期内最多只有一个Candidate能够赢得选举(这是选举的安全属性)。 
一旦Candidate赢得选举，它就成为Leader。 然后它将心跳消息发送给所有其它服务节点，从而建立其权限并阻止新的选举。

在等待投票结果时，Candidate可能会从另一个声称是Leader的服务节点收到AppendEntries RPC。 如果领导者的任期（包含在其RPC中）至少与
Candidate的当前时间周期一样大，那么Candidate将该Leader认为是合法的，并恢复到Follower状态。 
如果RPC消息中包含的时间周期小于Candidate的当前时间周期，则Candidate拒绝该RPC请求并继续处于选举状态。

第三种可能的结果是，Candidate既不会赢得选举，也不会输掉选举：如果很多Follower同时成为Candidate，投票可能会分裂，这样就没有Candidate能获得多数投票。
当发生这种情况时，每个Candidate都会超时，并通过增加时间周期的方式来启动新的选举，并发起另一轮的Request-Vote RPC请求。 
但是，如果没有额外措施，分裂的投票可能会无限期地重复。

Raft使用随机选举超时来确保分裂选票很少发生，或即使发生也能迅速解决。 首先，为了防止分裂投票，选举超时的时间间隔会从固定的时间间隔（例如150-300ms）中随机选择。 
这样扩展了服务器，使得在大多数情况下只有一个服务节点会超时;它赢得选举，并在其它所有服务节点超时之前发送心跳。同样的机制用于处理分裂投票。 
每个Candidate在选举开始时重新设置随机选举超时时间，并在开始下一次选举前等待超时。这样减少了新选举中再次分裂投票的可能性。









## 参考文献
https://github.com/tiglabs/raft

动画演示：http://thesecretlivesofdata.com/raft/
