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
prometheus --config.file="prometheus.yml"
```

## Node Exporter
```shell script
wget https://github.com/prometheus/node_exporter/releases/download/v0.18.1/node_exporter-0.18.1.darwin-amd64.tar.gz
```

# Prometheus HTTP API
**[HTTP API](https://prometheus.io/docs/prometheus/latest/querying/api/)**
