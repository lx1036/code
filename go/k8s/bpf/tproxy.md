

# tproxy

内核文档：https://www.kernel.org/doc/Documentation/networking/tproxy.txt

部署命令：

```

# 配置 tproxy iptables 规则把 包 redirect 到 127.0.0.1:10000, 正是 nginx stream server listen 10000.
iptables -t mangle -A PREROUTING -p tcp -m tcp -d 172.25.25.25/32 --dport 1:65499 -j TPROXY --on-port 10000 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0
iptables -t mangle -A PREROUTING -p udp -m udp -d 172.25.25.25/32 --dport 1:65499 -j TPROXY --on-port 10000 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0

# 2. 需要 listen server socket 支持 IP_TRANSPARENT，这样可以获取初始的 ip:port，而不是 port 10000
setsockopt(fd, SOL_IP, IP_TRANSPARENT, &value, sizeof(value));
```

