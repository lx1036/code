

# tcpdump
tcpdump 使用 ebpf 技术监听网络包

(1) tcpdump 监听容器内服务，查看 mtu
```shell
# nginx/busybox 容器内
curl -k https://36.102.10.242
# 另外窗口监听
tcpdump -nn -i any host 36.102.10.242
```

