

# MVCC 解决的问题
数据库领域中，解决高并发环境下数据冲突的问题。它的基本思想是保存一个数据的多个历史版本，从而解决事务管理中数据隔离的问题。



## 几个数据结构
treeIndex: 一个B+Tree，每一个节点是keyIndex，也就是由keyIndex组成的B+Tree。

revision: 每一次操作的逻辑时钟，{main, sub} main表示 transaction_id，sub 表示事务里的每一个操作id

keyIndex: 表示一个 key 有哪些 revision 的数据结构，包含最新的版本号 modified，以及历史版本号 generations(key在每一代中有哪些版本号)，这个最新版本号就很重要了。



## MVCC Watch Keys
watch 代码原理：https://time.geekbang.org/column/article/341060


## 参考文献

**[MVCC 在 etcd 中的实现](https://blog.betacat.io/post/mvcc-implementation-in-etcd/)**

**[etcd源码阅读与分析（五）：mvcc](https://jiajunhuang.com/articles/2018_11_28-etcd_source_code_analysis_mvvc.md.html)**

