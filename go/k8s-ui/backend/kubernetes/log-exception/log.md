
https://github.com/sirupsen/logrus

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


