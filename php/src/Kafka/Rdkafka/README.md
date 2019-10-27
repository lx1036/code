```shell script
# start in background
brew services zookeeper start
brew services kafka start

# stop
brew services kafka stop
brew services zookeeper stop

# 创建一个topic
kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic sunday
# 查看topic列表
kafka-topics --list --zookeeper localhost:2181
# 创建一个生产者
kafka-console-producer --broker-list localhost:9092 --topic sunday
# 创建一个消费者
kafka-console-consumer --bootstrap-server localhost:9092 --topic test --from-beginning
```
