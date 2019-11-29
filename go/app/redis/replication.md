
# 本地部署 redis replication

1. 在 `/usr/local/etc/redis/replication` 创建 `master/slave` 两个文件夹，并复制 `/usr/local/etc/redis.conf` 文件到两个文件夹内。
2. 修改 `master/redis.conf` 的 port 为 7700，`slave/redis.conf` 的 port 7701，和 `slaveof 127.0.0.1 7700`。
3. 运行 `redis-server ./master/redis.conf` 和 `redis-server ./slave/redis.conf`。如果需要后台运行，设置 `daemonize yes`。


# Sentinel 设置主从切换
