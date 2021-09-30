


# Backend
Backend 是一个接口，主要实现了 "事务批量写boltdb和buffer" 功能。
事务批量提交是 boltdb 非常核心的功能：https://github.com/etcd-io/bbolt#batch-read-write-transactions


## TroubleShooting
(1)etcd 一个写请求是如何执行的？
https://time.geekbang.org/column/article/336766
