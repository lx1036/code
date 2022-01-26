



## boltdb
boltdb 是一个 key-value 存储系统，麻雀虽小五脏俱全，最大好处：这样可以无需使用类似mysql这样庞大的数据库系统来存储数据。
关键功能：支持事务(批量写操作)，并发安全，range/prefix scan，数据备份，bucket分片

read-write transaction 会锁文件。
read-only transaction 没有锁文件，性能好。

## boltdb cli
```shell
go get github.com/br0xen/boltbrowser/...
go get go.etcd.io/bbolt/... # 安装bbolt命令，bolt cli客户端
```

### meta page

```shell
# 前两页是 meta page，至于为何有两个 meta page，主要是用来实现事务的!!!
bbolt page my.db 0
bbolt page my.db 1
```


## (1) boltdb 数据组织
* page:

* node:

boltdb 源码导读（一）：boltdb 数据组织: https://zhuanlan.zhihu.com/p/332439403


## (2) boltdb 索引结构
索引有两种数据结构：B-tree/B+tree 和 LSM-tree，LSM-tree 随机写性能更好，B-tree/B+tree 范围查询读性能更好。
所以 etcd 选择 B-tree, boltdb 选择 B+tree 作为索引数据结构，是有道理的。 

boltdb 源码导读（二）：boltdb 索引设计: https://zhuanlan.zhihu.com/p/341416264


## (3) boltdb 事务实现




## 参考文献
**[bbolt github](https://github.com/etcd-io/bbolt)**

**[阿里二面：什么是mmap？](https://zhuanlan.zhihu.com/p/357820303)**
