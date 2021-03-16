# prometheus docs
**[官方文档中文版](https://ryanyang.gitbook.io/prometheus/di-san-zhang-prometheus/storage)**

# Go client
**[prometheus go-client](https://github.com/prometheus/client_golang)**

# php client
**[itsmikej/prometheus_client_php_wrapper](https://github.com/itsmikej/prometheus_client_php_wrapper)**

# 安装方式
```shell script
wget -c https://github.com/prometheus/prometheus/releases/download/v2.14.0/prometheus-2.14.0.darwin-amd64.tar.gz
tar zxvf prometheus-2.14.0.darwin-amd64.tar.gz
mv prometheus /usr/local/bin/prometheus
```

# 运行
```shell script
prometheus --config.file="prometheus.yml" # http://localhost:9090,http://localhost:9090/metrics
```

# PromQL


# Prometheus 需要监控哪些对象？
对于一套Kubernetes集群而言，需要监控的对象大致可以分为以下几类：
* Kubernetes系统组件：Kubernetes内置的系统组件一般有apiserver、controller-manager、etcd、kubelet等，为了保证集群正常运行，我们需要实时知晓其当前的运行状态。
* 底层基础设施：Node节点(虚拟机或物理机)的资源状态、内核事件、CPU、内存使用率等。
* Kubernetes对象：主要是Kubernetes中的工作负载对象，如Deployment、DaemonSet、Pod等。
* 应用指标：应用内部需要关心的数据指标，如httpRequest。

# **[Prometheus 是什么？](https://docs.ucloud.cn/uk8s/monitor/prometheus/intro)**
Prometheus 整体架构:
![arch](./prometheus-architecture.png)

解释：主要模块是由Go写的Prometheus Server 模块，一个二进制文件，*收集和存储(使用时间序列数据库TSDB落地数据)时间序列数据(时间序列数据就是当前程序的实时运行状态)*。
同时比较重要的就是exporter模块(该模块代码作为API接口，供prometheus server调用，来抓取当前软件的实时状态)，以及
alert manager模块(作为报警使用，比如邮件、钉钉或者slack等等)。

## Prometheus 核心指标
metrics 都是类似 **prometheus_http_request_duration_seconds_bucket{handler="/-/healthy",le="60"} 408**
这样的形式写的，metric_name{label, ...} value, label 是 key=value 形式。总共有4种数据类型：
1. Counter 计数器
如请求的 QPS
2. Gauge 仪表盘
如温度计

3. Histogram 柱状图
如请求的延迟时间 P95/P99，貌似 Histogram 用的多些，Summary 用的少，有些客户端包如php的貌似不支持 Summary 的

4. Summary 类似 Histogram
如请求的延迟时间 P95/P99


# 如何部署 Prometheus
在 Docker 中安装：

在 K8S 中安装：**[部署Prometheus](https://docs.ucloud.cn/uk8s/monitor/prometheus/installprometheus)**


# 高可用 prometheus
使用 thanos 软件来实现 prometheus 高可用。
