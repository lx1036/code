
```shell script
ip link add veth0 type veth peer name eth0
ip netns add ns0
ip link set eth0 netns ns0
ip netns exec ns0 ip a add 10.20.1.2/24 dev eth0
ip netns exec ns0 ip link set eth0 up
ip netns exec ns0 ip route add 169.254.1.1 dev eth0 scope link
ip netns exec ns0 ip route add default via 169.254.1.1 dev eth0
ip link set veth0 up
ip route add 10.20.1.2 dev veth0 scope link
ip route add 10.20.1.3 via 192.168.1.16 dev ens192
echo 1 > /proc/sys/net/ipv4/conf/veth0/proxy_arp
```


