# Go client

**[prometheus go-client](https://github.com/prometheus/client_golang)**


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
