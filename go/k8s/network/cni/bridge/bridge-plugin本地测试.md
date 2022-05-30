
# 本地测试 bridge plugin
**[测试 CNI plugin](https://mp.weixin.qq.com/s/h7c1f18wY1NDFVUQcFD3ug)**

```shell
# 先下载 plugins
curl -O -L https://github.com/containernetworking/plugins/releases/download/v1.1.1/cni-plugins-linux-amd64-v1.1.1.tgz
tar zxf cni-plugins-linux-amd64-v1.1.1.tgz

cat > bridge.conf << "EOF"
{
    "cniVersion": "0.3.1",
    "name": "mybridge",
    "type": "bridge",
    "bridge": "cni_bridge0",
    "isDefaultGateway": true,
    "forceAddress": false,
    "ipMasq": true,
    "hairpinMode": true,
    "ipam": {
        "type": "host-local",
        "subnet": "100.100.100.0/24",
        "routes": [
            {"dst": "200.200.200.0/24", "gw": "100.100.100.1"}
        ]
    }
}
EOF

# 新建个 container netns
ip netns add netns-br-1
ip netns add netns-br-2
CNI_COMMAND=ADD CNI_CONTAINERID=container-id-1 CNI_NETNS=/var/run/netns/netns-br-1 CNI_IFNAME=eth0 CNI_PATH=`pwd` ./bridge < bridge.conf
CNI_COMMAND=ADD CNI_CONTAINERID=container-id-2 CNI_NETNS=/var/run/netns/netns-br-2 CNI_IFNAME=eth0 CNI_PATH=`pwd` ./bridge < bridge.conf

# 验证
# 查看 cni_bridge0 网卡 IP 100.100.100.1 和 路由, cni_bridge0 IP 为 100.100.100.1 作为网关
ip addr show cni_bridge0
61355: cni_bridge0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default qlen 1000
    link/ether 42:a5:98:38:c6:77 brd ff:ff:ff:ff:ff:ff
    inet 100.100.100.1/24 brd 100.100.100.255 scope global cni_bridge0
       valid_lft forever preferred_lft forever
ip route
100.100.100.0/24 dev cni_bridge0 proto kernel scope link src 100.100.100.1

ip netns exec netns-br-1 ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
2: eth0@if61358: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP group default
    link/ether a2:52:7d:1c:71:30 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 100.100.100.4/24 brd 100.100.100.255 scope global eth0
       valid_lft forever preferred_lft forever
ip netns exec netns-br-1 ip route
default via 100.100.100.1 dev eth0
100.100.100.0/24 dev eth0 proto kernel scope link src 100.100.100.4
200.200.200.0/24 via 100.100.100.1 dev eth0


iptables-save -t nat | grep mybridge
-A POSTROUTING -s 100.100.100.4/32 -m comment --comment "name: \"mybridge\" id: \"container-id-3\"" -j CNI-e566dfd049f2054d7313fa2c
-A CNI-e566dfd049f2054d7313fa2c -d 100.100.100.0/24 -m comment --comment "name: \"mybridge\" id: \"container-id-3\"" -j ACCEPT
-A CNI-e566dfd049f2054d7313fa2c ! -d 224.0.0.0/4 -m comment --comment "name: \"mybridge\" id: \"container-id-3\"" -j MASQUERADE

ip netns exec netns-br-3 ping 100.100.100.2
```


