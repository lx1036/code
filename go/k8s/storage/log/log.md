
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
