



## boltdb
boltdb 是一个 key-value 存储系统，麻雀虽小五脏俱全，最大好处：这样可以无需使用类似mysql这样庞大的数据库系统来存储数据。
关键功能：支持事务(批量写操作)，并发安全，range/prefix scan，数据备份，bucket分片

read-write transaction 会锁文件。
read-only transaction 没有锁文件，性能好。




## 参考文献
**[bbolt github](https://github.com/etcd-io/bbolt)**
