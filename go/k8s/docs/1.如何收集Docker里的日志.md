
# 如何收集Docker容器里的日志？
几种方式：
* 直接调用API，写入日志中心。具体为：调用第三方日志软件API，直接写入日志。
* 日志写入本地固定文件，然后通过日志采集器收集解析，再发送到日志中心。Pod里的Container把日志写入`/var/lib/docker/containers/*/*-json.log`文件内，
DaemonSet部署的日志收集器如Filebeat收集日志，发送到Kafka里，然后Logstash去消费日志，过滤聚合日志，再发送到日志存储中心ElasticSearch里，最后在Kibana里展示。
* 直接直接写入stdout/stderr。


 




## ElasticSearch



## Kibana







