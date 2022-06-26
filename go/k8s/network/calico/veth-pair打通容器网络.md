



(1) 配置 veth pair 和网卡 ip
```shell script
# 这个不完整，没有经过验证
ip link add veth0 type veth peer name eth0
ip netns add ns0
ip link set eth0 netns ns0
ip netns exec ns0 ip addr add 10.20.1.2/24 dev eth0
ip netns exec ns0 ip link set eth0 up
ip netns exec ns0 ip route add 169.254.1.1 dev eth0 scope link
ip netns exec ns0 ip route add default via 169.254.1.1 dev eth0
ip link set veth0 up
ip route add 10.20.1.2 dev veth0 scope link
ip route add 10.20.1.3 via 192.168.1.16 dev ens192
echo 1 > /proc/sys/net/ipv4/conf/veth0/proxy_arp
```

```shell
# veth pair 打通容器网络
# 可以看这个，已经经过验证
ip link add veth-test-2 type veth peer name veth-test-3
ip netns add net-veth-2
ip link set veth-test-2 netns net-veth-2
ip link set veth-test-2 up
ip netns exec net-veth-2 ip link set veth-test-2 up
ip netns exec net-veth-2 ip route add 169.254.1.1 dev veth-test-2
ip netns exec net-veth-2 ip route add default via 169.254.1.1 dev veth-test-2
ip netns exec net-veth-2 ip neigh add 169.254.1.1 dev veth-test-2 lladdr ee:ee:ee:ee:ee:ee
ip link set addr ee:ee:ee:ee:ee:ee veth-test-3
ip netns exec net-veth-2 ip addr add 100.162.253.162 dev veth-test-2 # 100.162.253.162 随便写的 ip
ip route add 100.162.253.162 dev veth-test-3
ip netns exec net-veth-2 curl -I 192.168.246.174 # 192.168.246.174 为 service ip
```
