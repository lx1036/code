

# Problem
如何在k8s环境中，搭建log架构，每一个container如何去接入log。

# Solution
搭建一套EFK(ElasticSearch, Fluentd, Kibana)，每个容器的日志输出为stdout标准输出，同时也可以添加EFK hook。这样同时写入两个地方。

# Handlers
* stdout
* file
* syslog
* efk
* sentry(exception)


# PHP Solution


# Golang Solution
使用 **[logrus hook](https://github.com/sirupsen/logrus#hooks)** 设置多个 log handler：stdout 和 EFK。


```shell
curl -O https://artifacts.elastic.co/downloads/kibana/kibana-7.6.1-darwin-x86_64.tar.gz
curl https://artifacts.elastic.co/downloads/kibana/kibana-7.6.1-darwin-x86_64.tar.gz.sha512 | shasum -a 512 -c -
tar -xzf kibana-7.6.1-darwin-x86_64.tar.gz
cd kibana-7.6.1-darwin-x86_64/

curl -L -O https://artifacts.elastic.co/downloads/beats/filebeat/filebeat-7.6.1-darwin-x86_64.tar.gz
tar xzvf filebeat-7.6.1-darwin-x86_64.tar.gz
cd filebeat-7.6.1-darwin-x86_64/

wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.6.1-darwin-x86_64.tar.gz
wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.6.1-darwin-x86_64.tar.gz.sha512
shasum -a 512 -c elasticsearch-7.6.1-darwin-x86_64.tar.gz.sha512
tar -xzf elasticsearch-7.6.1-darwin-x86_64.tar.gz
cd elasticsearch-7.6.1/

```


# logrotate
**[Linux 日志切割神器 logrotate 原理介绍和配置详解](https://wsgzao.github.io/post/logrotate/)**

logrotate 是一个 linux 系统日志的管理工具。
可以对单个日志文件或者某个目录下的文件按时间 / 大小进行切割，压缩操作；指定日志保存数量；还可以在切割之后运行自定义命令。
logrotate 是基于 crontab 运行的，所以这个时间点是由 crontab 控制的，具体可以查询 crontab 的配置文件 /etc/anacrontab。
系统会按照计划的频率运行 logrotate，通常是每天。在大多数的 Linux 发行版本上，计划每天运行的脚本位于 /etc/cron.daily/logrotate。
如果找不到，可以 `apt/yum install -y logrotate` 。
logrotate的配置文件：
```shell script
# logrotate 定时任务配置
sudo cat /etc/cron.daily/logrotate
# logrotate 配置
sudo cat /etc/logrotate.conf
```


# filebeat
filebeat 解决的几个重要问题：
* 日志断点续传：这个问题很重要，filebeat 收割日志到某一行之后，filebeat 重启之后可以借助 registry state offset 继续从断点开始继续
发到 output 模块中

## 参考文献
**[Elastic-Filebeat 实现原理剖析](https://www.cyhone.com/articles/analysis-of-filebeat/)**
**[监控日志系列-Filebeat原理](https://kingjcy.github.io/post/monitor/log/collect/filebeat/filebeat-principle/)**
