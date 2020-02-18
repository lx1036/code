# Download
```shell script
wget https://github.com/prometheus/node_exporter/releases/download/v0.18.1/node_exporter-0.18.1.darwin-amd64.tar.gz
tar -xzf node_exporter-0.15.2.darwin-amd64.tar.gz
mv node_exporter-0.18.1.darwin-amd64/node_exporter /usr/local/bin/node_exporter
node_exporter
```

# 查看node的监控指标
```shell script
prometheus --config.file="prometheus.yml"
node_exporter # http://localhost:9100
```
