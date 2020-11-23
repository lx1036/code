
# Filebeat add_kubernetes_metadata 插件源码解析
用户使用指南：https://www.elastic.co/guide/en/beats/filebeat/master/add-kubernetes-metadata.html
字段配置文档：https://www.elastic.co/guide/en/beats/filebeat/current/exported-fields-kubernetes-processor.html

## 背景
该文章分析filebeat版本基于7.10。
部署在k8s上业务pod内各个容器日志可以输出到stdout/stderr，docker的log drivers比如json-file，会把日志根据64位容器id分类，以json形式默认存入
路径下`/var/lib/docker/containers/`：

![docker_logs_path](./imgs/docker_logs_path.png)

可以修改`/etc/docker/daemon.json`里的`graph`配置修改日志存储路径。

filebeat的作用是消费这些日志文件，并存入第三方工具如kafka，供后续进一步消费；或者方便测试，直接终端输出。filebeat的标准配置文件类似如下：
```yaml
# filebeat.yaml文件内容如下
filebeat.config.inputs:
  enabled: true
  path: ${path.config}/inputs.yml
  reload.enabled: true
  reload.period: 10s
processors:
  - add_kubernetes_metadata:
    host: slave01
    kube_config: /Users/liuxiang/.kube/config
    in_cluster: true
    matchers:
      - logs_path:
          logs_path: /Users/liuxiang/Code/k8s/beats/filebeat

output.console:
  pretty: true

#output.kafka:
#  hosts: ["127.0.0.1"]
#  topic: "test"

# input.yaml文件内容如下
- type: log
  paths:
    - /Users/liuxiang/Code/k8s/beats/filebeat/51d52e9b3645a9f7d7149d471335ca45bb547d87625999d03b46e252c700a505-json.log
```

## 目的
学习编写filebeat plugin。
学习二次开发k8s。
学习golang。

## add_kubernetes_metadata 插件用途
主要是给每一行日志添加一些k8s相关的元数据metadata，比如namespace、deployment/pod name、labels等等。


## 基本概念


### Indexer


### Matcher




## 本地测试



## 工作原理
add_kubernetes_metadata启动时






## 参考文献

