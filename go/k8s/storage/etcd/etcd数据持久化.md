
**![架构](./imgs/etcd-arch.png)**

一个请求处理过程：

etcd-server -> etcd-raft -> wal
etcd使用mvcc多版本并发控制乐观锁来解决并发冲突问题：处理写请求时，mvcc会为这一key添加version(mainID,subID)新版本，并写入新数据；
读操作时，读的是之前版本旧数据，使得读写分离，读写不会阻塞。所以etcd特别适合多读场景。


# 数据持久化
(1)Etcd是如何实现数据持久化的？
wal(预写式日志write-ahead-log，类似binlog，所有的修改在提交之前都要写入log文件)+snapshot方式实现持久化。

**[Etcd存储的实现](https://www.codedump.info/post/20181125-etcd-server/)**
**[etcd源码阅读与分析（五）：mvcc](https://jiajunhuang.com/articles/2018_11_28-etcd_source_code_analysis_mvvc.md.html)**: mvcc底层使用 bolt 实现，bolt是一个基于B+树的KV存储。
在数据库领域，面对高并发环境下数据冲突的问题，业界常用的解决方案有两种(**[MVCC 在 etcd 中的实现](https://blog.betacat.io/post/mvcc-implementation-in-etcd/)**):
* 悲观锁
* 乐观锁，如MVCC（Multi-version Concurrent Control）

etcd存储层可以看成由两部分组成，一层在内存中的基于btree的索引层，一层基于boltdb的磁盘存储层。

(2)BTree实现？
https://github.com/google/btree


(3)etcd 是如何用 bbolt 来存储 key-value？


(4)etcd 如何保证数据读写的事务性？


# WAL文件如何实现

