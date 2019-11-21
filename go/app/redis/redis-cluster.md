
# 使用教程
**[Redis cluster tutorial](https://redis.io/topics/cluster-tutorial)**

Redis Cluster provides a way to run a Redis installation where data is **automatically sharded across multiple Redis nodes**.

Redis Cluster 的作用：
* 可以分割数据集到多个 redis nodes。
* 高可用，只有当**多个** master redis nodes 不可用时，集群才不可用。 

## Redis Cluster TCP ports
Redis Cluster 的每一个 node 都会有两个 tcp port 与其他节点连接，如 ports 7000/17000（7000 port 是与 redis client 连接并传数据，17000 port 是与其他 redis node 连接并传输数据）。
Redis node 之间通信使用的是 Redis 自定义的二进制协议。

### redis cluster bus
内部 redis nodes 之间通信

### Redis Cluster 与 Docker 
这里有网络的问题：

TODO: 
Docker network 的文档：**[docker network](https://docs.docker.com/network/)**

### Redis Cluster Data Sharding
Redis Cluster 没有使用**一致性哈希 consistent hashing**来做数据分片，而是使用了**哈希槽 hash slot**。
什么是**一致性哈希 consistent hashing**？

一个 redis cluster 总共有 16384=2^14 hash slot，每一个 redis node 平均分配 2^14 个 hash slot。
key 落入 hash slot 的算法如下：
```shell script
hash_slot = CRC16(key)
```

那添加和删除 cluster node 时，hash slot 会如何分配？
如已有 A/B/C 3个 redis node，添加一个 D，只会最少移动 hash slot。**Redis 从一个 node 移动 hash slot 到另一个 node，不会阻塞 redis crud 操作。**

#### hash tags
为了让多个 key 落入一个 hash slot 内，引入 hash tag，如 this{foo}key/that{foo}key 这两个 key，redis 只会 CRC16 {}内的字符串，所以这两个 key 会落入同一个 hash slot：
```shell script
// this{foo}key/that{foo}key
hash_slot = CRC16(foo)
```

## Redis Cluster master-replicas
为了高可用，redis cluster 设计了每一个 master node 有多个副本节点 replica node，如果 A master node 挂了，会从其副本节点中选择一个作为 master node(这里 Etcd 使用 Raft 算法投票选举 master node)。

Redis Cluster 不能保证强一致性，比如写入操作时，内部会发生：
* Client 写数据 test1 => value1，根据 CRC16(test1) 落入 B master node 内的一个 hash slot。
* B master node 写入成功后会立即返回 OK 给 Client。
* **然后 B master node 再把写数据命令，通过 cluster bus 传给其多个副本节点 B1,B2...**。

为了提高延迟追求性能，这里 redis cluster 没有先把写数据命令传给副本节点，等待多个副本节点给予确认后，最后再返回 OK 给 Client。
所以，Redis Cluster 不能保证强一致性，而 Etcd 可以保证强一致性（使用 Raft 算法），Etcd 就是等待多个副本节点给予确认后，最后再返回 OK 给 Client。

## Redis Cluster Configuration Parameters
* cluster-enabled
* cluster-config-file
* cluster-node-timeout
* cluster-slave-validity-factor
* cluster-migration-barrier
* cluster-require-full-coverage


有了 6 个 cluster nodes 后，使用 redis-cli 创建一个 cluster：
```shell script
redis-cli --cluster create 127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 --cluster-replicas 1
```
**--cluster-replicas 1** 表示为每一个 master node 创建一个 replica node。 

### Redis Cluster Client
* 1. `redis-cli -c -p port` 使用 `-c` 切换成与 cluster 交互模式，不带 `-c` 是与单个 redis instance 交互。
* 2. **[go-redis](https://github.com/go-redis/redis)** 是 Go 写的 redis 客户端，包含与 cluster 交互功能。


## Reshard Redis Cluster
```shell script
redis-cli -p 7000 cluster nodes | grep myself
redis-cli --cluster reshard 127.0.0.1:7000
redis-cli --cluster check 127.0.0.1:7000
```
或者 
```shell script
redis-cli reshard <host>:<port> --cluster-from <node-id> --cluster-to <node-id> --cluster-slots <number of slots> --cluster-yes
```

## Delete a master node


## Add a new master node

### Add a new replica node



## Replicas migration


## Upgrading nodes in cluster



# 设计文档
**[Redis Cluster Specification](https://redis.io/topics/cluster-spec)**
