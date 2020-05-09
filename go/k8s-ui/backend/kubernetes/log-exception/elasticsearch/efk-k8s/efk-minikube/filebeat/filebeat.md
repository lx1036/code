

# **[Filebeat](https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-overview.html)**
Filebeat 从配置文件中的 inputs(数据源，如/var/lib/docker/containers/*/*.log)读取数据，每一个log文件数据
都会起一个harvester(数据收割机)去读取数据，然后聚合发送数据到outputs(如 elasticsearch)。原理图如下：

![filebeat](./filebeat.png)

## **[配置 filebeat](https://www.elastic.co/guide/en/beats/filebeat/current/filebeat-configuration.html)**




