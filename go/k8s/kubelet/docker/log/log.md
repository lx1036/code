
支持的log-driver列表：https://docs.docker.com/config/containers/logging/configure/#supported-logging-drivers
自定义log driver plugin官方文档：https://github.com/docker/cli/blob/master/docs/extend/plugins_logging.md，
参考示例：https://github.com/cpuguy83/docker-log-driver-test


## **[日志log](https://docs.docker.com/config/containers/logging/)**
一般通过`docker logs $(container_id)`查看日志，或者`/var/lib/docker/containers/*`里找log日志，日志会写入`/dev/stdout或/dev/stderr`。
Docker的**[logging driver](https://docs.docker.com/config/containers/logging/configure/)**会把`stdout/stderr`写到
`/var/lib/docker/containers/$(container_id)/$(container_id)-json.log`文件内，且默认使用`json-file` log driver，
可以通过`docker inspect -f '{{.HostConfig.LogConfig.Type}}' <Container>`查看如`docker inspect -f '{{.HostConfig.LogConfig.Type}}' filebeat`
查看其它驱动命令：
```shell script
docker info --format '{{.Plugins.Log}}'
# https://docs.docker.com/config/containers/logging/configure/#supported-logging-drivers
# [awslogs fluentd gcplogs gelf journald json-file local logentries splunk syslog]

docker info -f '{{.LoggingDriver}}'
# json-file
```
其中，`fluentd`会直接把日志发送给**[Fluentd服务](http://www.fluentd.org)**，再把日志发送给Kafka -> Logstash -> ES。

**[Docker双栈日志 duel logging](https://mp.weixin.qq.com/s/oZ5xbCbO_1lsgEa3QKBxoQ)**：
## Demo:以fluentd作为log driver
(1) 启动 Fluent Bit 容器(log collector)来收集容器日志：
```shell script
docker run -p 24224:24224 -v $(pwd)/log/fluentd/docker_to_stdout.conf:/docker_to_stdout.conf -d --name=fluentd fluent/fluent-bit:1.3
```

(2) 启动一个log driver为fluentd的容器，并容器内打印日志到stdout
```shell script
# **[Customize log driver output](https://docs.docker.com/config/containers/logging/log_tags/)**
docker run --log-driver=fluentd --log-opt fluentd-address=tcp://localhost:24224 --log-opt tag="[fluentd]" alpine echo "hello world"
```

(3)可以进去任意一个容器内`docker exec -it filebeat /bin/bash`，在`/var/lib/docker/containers/$(container_id)/$(container_id)-json.log`文件内
去`cat *.log | grep "hello world"`查找日志。并且`docker logs $(docker ps -ql)`就会显示：
```markdown
Error response from daemon: configured logging driver does not support reading
```

(4) 问题来了，如果使用非`local`或`json-file`或`journald`驱动的log driver，就不能在本地使用`dockers log $(container)`来查看日志了。
解决办法：使用docker双栈日志。`Docker CE 20.03 里增加了双栈日志功能，会在*/var/lib/docker/containers/$(container_id)/*目录内增加一个container-cached.log文件,
缓存容器的日志。`



# 写一个自定义的log driver plugin
**[Docker log driver plugins](https://docs.docker.com/engine/extend/plugins_logging/)**

# 安装一个Docker Plugin
```shell script
docker plugin install lx1036/log-driver:1.0.0
docker plugin ls
```
