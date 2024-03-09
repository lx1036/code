

# 1. 安装 tproxy iptables 规则

eth0_ip=`ip addr show eth0 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1`
echo $eth0_ip

iptables -t mangle -A PREROUTING -p tcp -d $eth0_ip/32 --dport 8000:8999 -j TPROXY --on-port 9090 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0
iptables -t mangle -A PREROUTING -p tcp -d $eth0_ip/32 --dport 7000:7999 -j TPROXY --on-port 9091 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0
iptables -t mangle -A PREROUTING -p udp -d $eth0_ip/32 --dport 6000:6999 -j TPROXY --on-port 9092 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0

# 172.16.0.154/32
# iptables -t mangle -A PREROUTING -p udp -d $eth0_ip/32 --dport 1000:2000 -j TPROXY --on-port 2001 --on-ip 127.0.0.1 --tproxy-mark 0x0/0x0 -m comment --comment "test for udp transparent"
# iptables -t mangle -S
# iptables -t mangle -D PREROUTING 1


# 2. 启动 nginx 容器
docker stop ga-monitor && docker rm ga-monitor
docker run --network host --name ga-monitor -v /root/nginx/nginx-tproxy.conf:/etc/nginx/nginx.conf:ro -v /root/nginx/index.html:/usr/share/nginx/html/index.html -d lx1036/nginx:1.25.2-transparent

