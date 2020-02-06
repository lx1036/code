

**[Etcd 文档](https://etcd.io/docs/v3.4.0/integrations/)**

**[Etcd go client 包](https://github.com/etcd-io/etcd/blob/master/clientv3/README.md)**

```shell script
brew install goreman
goreman -f Procfile start # 启动 etcd 集群

etcdctl --endpoints=localhost:12379 --write-out=table member list # 获取集群 member 信息
etcdctl --endpoints=localhost:12379 put foo bar
etcdctl --endpoints=localhost:22379 get foo

# 检测 etcd 容灾能力
goreman run stop etcd1
etcdctl --endpoints=localhost:22379 put foo1 bar1
etcdctl --endpoints=localhost:32379 get foo1
etcdctl --endpoints=localhost:12379 get foo1 # 强制链接，报错 'DeadlineExceeded'

goreman run restart etcd1
etcdctl --endpoints=localhost:12379 get foo1 # 重启，关闭期间的数据会重新恢复
```



## Etcd 文章列表
0. **[Etcd 中文文档](https://doczhcn.gitbook.io/etcd/)**
1. **[Raft算法原理](https://www.codedump.info/post/20180921-raft/)**
2. **[etcd Raft库解析](https://www.codedump.info/post/20180922-etcd-raft/)**
3. **[Etcd存储的实现](https://www.codedump.info/post/20181125-etcd-server/)**
