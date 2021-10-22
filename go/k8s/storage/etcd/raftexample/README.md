

# raft example
https://github.com/etcd-io/etcd/blob/main/contrib/raftexample/README.md
提供了一个 HTTP Server Rest API，并借助 KVStore，来 Propose pb.Message 到 raft state machine，
同时 KVStore 从 raft ReadyChan 获取数据并保存。
