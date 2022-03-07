
# net-tools

## tcpdump
tcpdump 使用 ebpf 技术监听网络包

(1) tcpdump 监听容器内服务，查看 mtu
```shell
# nginx/busybox 容器内
curl -k https://36.102.10.242
# 另外窗口监听
tcpdump -nn -i any host 36.102.10.242
# 抓包arp协议
tcpdump -i eth0 -nnee arp and host 20.206.230.25
```


## nsenter

```shell
# Example: nsenter-ctn <ctn-id> -n ip addr show eth0
function nsenter-ctn () {
    CTN=$1 # Container ID or name
    PID=$(sudo docker inspect --format "{{.State.Pid}}" $CTN)
    shift 1 # Remove the first argument, shift remaining ones to the left
    sudo nsenter -t $PID $@
}

# Put it into your ~/.bashrc then
source ~/.bashrc
```


## arp
```shell
apt install -y arping
docker run --name ctn1 -d alpine:3.12.0 sleep 30d
docker run --name ctn2 -d alpine:3.12.0 sh -c 'while true; do echo -e "HTTP/1.0 200 OK\r\n\r\nWelcome" | nc -l -p 80; done'

# 两个容器内抓包 arp
nsenter-ctn ctn1 -n arping 172.18.0.3
nsenter-ctn ctn2 -n tcpdump -i eth0 arp # 抓包 arp 包
nsenter-ctn ctn1 -n ping 172.18.0.3 -c 2
nsenter-ctn ctn2 -n tcpdump -i eth0 icmp # 抓包 ping 包

nsenter-ctn ctn1 -n curl 172.18.0.3 # 抓包 tcp 包, 
nsenter-ctn ctn2 -n tcpdump -i eth0 tcp -nnee
# 三次握手
# client -> server [Seq, seq=x]
# server -> client [Seq, ack=x+1, seq=y]
# client -> server [ack=1]
# 四次挥手, 客户端或服务端均可主动发起挥手动作，这里是 server 发起挥手动作
# server -> client [Fin, seq=x, ack=1]
# client -> server [ack=x]
# client -> server [Fin, seq=x]
# server -> client [ack=x+1]
```


# 参考文献
**[bpf arp demo](https://arthurchiao.art/blog/firewalling-with-bpf-xdp/)**
