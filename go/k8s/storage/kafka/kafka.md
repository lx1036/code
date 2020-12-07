
# kafka
kafka 是由 LinkedIn 开发的一个分布式的消息中间件。

## 安装Kafka并启动
```shell script
# kafka自带zookeeper，不需要单独安装zookeeper
# 配置文件在 /usr/local/etc/kafka/server.properties
# 找到 listeners=PLAINTEXT://:9092 那一行，把注释取消掉，然后重启
brew install kafka
brew services start/stop kafka
```

临时启动kafka:
```shell
zkServer start # 临时启动zookeeper
kafka-server-start /usr/local/etc/kafka/server.properties # 临时启动kafka
```

测试：
```shell
# 创建topic
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic test

```

## kafka cli/gui
**[fgeller/kt](https://github.com/fgeller/kt)**:
```shell script
brew tap fgeller/tap
brew install kt
```
gui:
**[conduktor](https://www.conduktor.io/)**

idea plugin: kafkalytic
