
# 面试问题：说一下 etcd mvcc 原理？
etcd=raft + mvcc
mvcc=treeIndex(索引) + boltdb(存储)

mvcc 包括读写事务两个过程：
写事务：构造该 key 的 revision，然后写入到 treeIndex btree 里，同时 buffer 内存写一份，buffer 是为了加速读事务。同时，写事务是批量 commit 到
boltdb 里，会起一个 goroutine 周期性提交。写事务需要拿到全局读写锁。写事务不支持并发写。
key 的每一个 revision 存储在 keyIndex.generation 结构体里，如果删除该 key，则会打上一个 tombstone 标记。

读事务：读事务只需要拿到全局读锁。查询 treeIndex btree 获得 keyIndex 里的 revision，然后根据该 revision 从 boltdb 里查询 value，该
value 包含真正的 key-value 值。读事务不会阻塞写事务，可以并发读。

同时，也补充聊下 mvcc watch key 功能，该功能是 K8S 的基石!!!
每一个 watch key 都会实例化一个 watcher 对象，然后一个 loop 去 reconcile 这个 watcher，当 key 写操作时，会同时获取这个 key 的所有
watcher 对象，然后把增量数据构造一个 WatchEvent 依次发给所有的 watcher。在 mvcc 上层构造了一个 grpcWatchServer，然后发给 grpcWatchClient。
同时，考虑到不会每次都能发送成功，发送失败的就暂放在 victim 结构体里，等下一次 loop 再去发送，让该 watcher 尽快追上进度。

