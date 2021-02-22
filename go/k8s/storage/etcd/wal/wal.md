

# Wal 定义
预写式日志wiki定义: https://zh.wikipedia.org/wiki/%E9%A2%84%E5%86%99%E5%BC%8F%E6%97%A5%E5%BF%97

源码中文注释：https://github.com/lichuang/etcd-3.1.10-codedump/blob/master/wal/wal.go



* wal：用于存放预写式日志，其最大的作用是记录整个数据变化的全部历程。在 Etcd 中，所有数据的修改在提交前，都要先写入 WAL 中。
使用 WAL 进行数据的存储使得 Etcd 拥有故障快速回复和数据回滚这两个重要功能。

* snap：用于存放快照数据。Etcd 为防止 WAL 文件过多会创建快照，snap 用于存储 Etcd 的快照数据状态。
