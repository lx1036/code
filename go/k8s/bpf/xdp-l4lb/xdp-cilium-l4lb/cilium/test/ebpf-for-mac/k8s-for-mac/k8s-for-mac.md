



# 1. 抓包 worker-eth0-udp

* 在 worker 上抓包 worker-eth0-udp.pcap
* 在 master 上 curl nginx pod ip

```shell
minikube ssh m02
sudo tcpdump -i eth0 -nneevv -A udp -w worker-eth0-udp.pcap

curl 10.244.1.3
```


# 2. 抓包 worker-cni0-tcp

* 在 worker 上抓包 worker-cni0-tcp.pcap
* 在 master 上 curl nginx pod ip

```shell
minikube ssh m02
sudo tcpdump -i cni0 -nneevv -A tcp -w worker-cni0-tcp.pcap

curl 10.244.1.3

minikube cp minikube-m02:/home/docker/worker-cni0-tcp.pcap worker-cni0-tcp.pcap
```



# 在 ecs 上 minikube 安装一个 k8s 集群，如何本地 lens 访问这个 k8s?
注意：本机没法 exec 到 pod 里。
1.安装 calico(v3.26.3) k8s
minikube stop && minikube delete
minikube start --cni=calico --driver=docker --kubernetes-version=v1.28.3 --force --listen-address=0.0.0.0
minikube node add --worker=true
2. ecs 上开启一个 proxy server, "172.16.3.161" 是 eth0 的 ip，还必须指定 --accept-hosts, 默认端口是 8001
kubectl proxy --address="172.16.3.161" --accept-hosts='^.*' --port=7007
3. 配置本地 kubeconfig
```markdown
本机的 profiles/minikube/ca.crt -> ecs 上的 ~/.minikube/profiles/minikube/ca.crt
本机的 profiles/minikube/client.crt -> ecs 上的 ~/.minikube/profiles/minikube/client.crt
本机的 profiles/minikube/client.key -> ecs 上的 ~/.minikube/profiles/minikube/client.key
server 写: http://${公网EIP}:7007
```

## 开启 calico bpf 模式
修改 default FelixConfiguration 资源对象：
```markdown
spec:
  bpfLogLevel: ''
  floatingIPs: Disabled
  logSeverityScreen: Info
  reportingInterval: 0s
  bpfKubeProxyIptablesCleanupEnabled: false # 不要清理 kube-proxy iptables 规则，和 kube-proxy 共同运行
  bpfEnabled: true
  bpfExternalServiceMode: Tunnel
  bpfConnectTimeLoadBalancingEnabled: false # 关闭这个 bpf 配置，否则 ecs 里跑，会报错
```

## ebpf 挂载
```shell
docker@minikube:~$ ip addr
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
2: tunl0@NONE: <NOARP,UP,LOWER_UP> mtu 1480 qdisc noqueue state UNKNOWN group default qlen 1000
    link/ipip 0.0.0.0 brd 0.0.0.0
    inet 10.244.120.66/32 scope global tunl0
       valid_lft forever preferred_lft forever
3: docker0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default 
    link/ether 02:42:fe:8f:8e:06 brd ff:ff:ff:ff:ff:ff
    inet 172.17.0.1/16 brd 172.17.255.255 scope global docker0
       valid_lft forever preferred_lft forever
5: calid478c22453f@if4: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether ee:ee:ee:ee:ee:ee brd ff:ff:ff:ff:ff:ff link-netnsid 2
9: cali80c0a5ec8c6@if4: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1480 qdisc noqueue state UP group default qlen 1000
    link/ether ee:ee:ee:ee:ee:ee brd ff:ff:ff:ff:ff:ff link-netnsid 1
16: eth0@if17: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 02:42:c0:a8:31:02 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 192.168.49.2/24 brd 192.168.49.255 scope global eth0
       valid_lft forever preferred_lft forever
docker@minikube:~$ tc filter show dev eth0 ingress
filter protocol all pref 49152 bpf chain 0
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_from_hos:[7093] direct-action not_in_hw id 7093 tag a73ddb0fcb16bcab 
docker@minikube:~$ tc filter show dev eth0 egress
filter protocol all pref 49152 bpf chain 0 
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_to_host_:[7098] direct-action not_in_hw id 7098 tag 537a96f4a7660ce4 
docker@minikube:~$ tc filter show dev calid478c22453f ingress
filter protocol all pref 49152 bpf chain 0 
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_from_wor:[7121] direct-action not_in_hw id 7121 tag 28f865493bf9c49f 
docker@minikube:~$ tc filter show dev calid478c22453f egress
filter protocol all pref 49152 bpf chain 0 
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_to_workl:[7119] direct-action not_in_hw id 7119 tag 5d61ef4f4b4d918f 
docker@minikube:~$ tc filter show dev tunl0 ingress
filter protocol all pref 49152 bpf chain 0 
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_from_l3d:[7097] direct-action not_in_hw id 7097 tag e2c945bb3a7c2a4a 
docker@minikube:~$ tc filter show dev tunl0 egress
filter protocol all pref 49152 bpf chain 0 
filter protocol all pref 49152 bpf chain 0 handle 0x1 calico_to_l3dev:[7096] direct-action not_in_hw id 7096 tag e53a1ce4d895726b 
docker@minikube:~$ 
```

## 跨节点 pod 访问
前提：calico tunnel 使用 ipip 协议来封包。
node1: 192.168.49.2
node2: 192.168.49.3
node3: 192.168.49.4

### node1 访问 node2-pod2
在 node2 上抓包 eth0 网卡，内层 ip header 源地址是 node1 的 ipip 网卡 tunl0 地址，外层 ip 地址是 node1 > node2 ip 地址。这是因为添加了路由：
```shell
# 路由解释：目的地址在 10.244.151.0/26 网段内，通过 tunl0 网卡连接邻居节点 192.168.49.4 进行的，路由协议是 bird 路由器软件维护的，onlink 表示可以直接访问全局IPv6网络。
10.244.151.0/26 via 192.168.49.4 dev tunl0 proto bird onlink
10.244.205.192/26 via 192.168.49.3 dev tunl0 proto bird onlink
```

抓包：
```shell
# node1 上访问 pod2 ip
docker@minikube:~$ curl 10.244.205.196 -v

docker@minikube-m02:~$ sudo tcpdump -i eth0 -nneevv proto 4
tcpdump: listening on eth0, link-type EN10MB (Ethernet), snapshot length 262144 bytes
06:27:54.920472 02:42:c0:a8:31:02 > 02:42:c0:a8:31:03, ethertype IPv4 (0x0800), length 94: (tos 0x0, ttl 64, id 29119, offset 0, flags [DF], proto IPIP (4), length 80)
    192.168.49.2 > 192.168.49.3: (tos 0x0, ttl 64, id 65227, offset 0, flags [DF], proto TCP (6), length 60)
    10.244.120.66.35448 > 10.244.205.196.80: Flags [S], cksum 0x5c1d (incorrect -> 0x6d81), seq 1325882737, win 64800, options [mss 1440,sackOK,TS val 2055458750 ecr 0,nop,wscale 7], length 0
06:27:54.920553 02:42:c0:a8:31:03 > 02:42:c0:a8:31:02, ethertype IPv4 (0x0800), length 94: (tos 0x0, ttl 63, id 19806, offset 0, flags [DF], proto IPIP (4), length 80)
    192.168.49.3 > 192.168.49.2: (tos 0x0, ttl 63, id 0, offset 0, flags [DF], proto TCP (6), length 60)
    10.244.205.196.80 > 10.244.120.66.35448: Flags [S.], cksum 0x5c1d (incorrect -> 0x09ca), seq 2028252921, ack 1325882738, win 64260, options [mss 1440,sackOK,TS val 15744244 ecr 2055458750,nop,wscale 7], length 0
06:27:54.920587 02:42:c0:a8:31:02 > 02:42:c0:a8:31:03, ethertype IPv4 (0x0800), length 86: (tos 0x0, ttl 64, id 29120, offset 0, flags [DF], proto IPIP (4), length 72)
    192.168.49.2 > 192.168.49.3: (tos 0x0, ttl 64, id 65228, offset 0, flags [DF], proto TCP (6), length 52)
    10.244.120.66.35448 > 10.244.205.196.80: Flags [.], cksum 0x5c15 (incorrect -> 0x318c), seq 1, ack 1, win 507, options [nop,nop,TS val 2055458750 ecr 15744244], length 0
```

### node1-pod1 访问 node2-pod2
在 node2 上抓包 eth0 网卡，内层 ip header 源地址是 pod1 ip 地址(10.244.120.68)，外层 ip 地址是 node1 > node2 ip 地址：
```shell
# pod1 里访问 pod2
docker@minikube:~$ sudo nsenter -t 43213 -n curl 10.244.205.196 -v

docker@minikube-m02:~$ sudo tcpdump -i eth0 -nneevv proto 4
06:32:17.551775 02:42:c0:a8:31:02 > 02:42:c0:a8:31:03, ethertype IPv4 (0x0800), length 94: (tos 0x0, ttl 63, id 57102, offset 0, flags [DF], proto IPIP (4), length 80)
    192.168.49.2 > 192.168.49.3: (tos 0x0, ttl 63, id 43686, offset 0, flags [DF], proto TCP (6), length 60)
    10.244.120.68.41652 > 10.244.205.196.80: Flags [S], cksum 0x5c1f (incorrect -> 0x7bb8), seq 658074349, win 64800, options [mss 1440,sackOK,TS val 2357506458 ecr 0,nop,wscale 7], length 0
06:32:17.551881 02:42:c0:a8:31:03 > 02:42:c0:a8:31:02, ethertype IPv4 (0x0800), length 94: (tos 0x0, ttl 63, id 61719, offset 0, flags [DF], proto IPIP (4), length 80)
    192.168.49.3 > 192.168.49.2: (tos 0x0, ttl 63, id 0, offset 0, flags [DF], proto TCP (6), length 60)
    10.244.205.196.80 > 10.244.120.68.41652: Flags [S.], cksum 0x5c1f (incorrect -> 0x7c27), seq 3943526790, ack 658074350, win 64260, options [mss 1440,sackOK,TS val 1349145757 ecr 2357506458,nop,wscale 7], length 0
06:32:17.551947 02:42:c0:a8:31:02 > 02:42:c0:a8:31:03, ethertype IPv4 (0x0800), length 86: (tos 0x0, ttl 63, id 57103, offset 0, flags [DF], proto IPIP (4), length 72)
    192.168.49.2 > 192.168.49.3: (tos 0x0, ttl 63, id 43687, offset 0, flags [DF], proto TCP (6), length 52)
    10.244.120.68.41652 > 10.244.205.196.80: Flags [.], cksum 0x5c17 (incorrect -> 0xa3e9), seq 1, ack 1, win 507, options [nop,nop,TS val 2357506458 ecr 1349145757], length 0
```
