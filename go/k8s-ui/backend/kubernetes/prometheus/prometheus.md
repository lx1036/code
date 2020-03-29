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
