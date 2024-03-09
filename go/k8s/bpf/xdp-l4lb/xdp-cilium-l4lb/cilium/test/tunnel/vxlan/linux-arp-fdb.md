

# 手动维护 arp/fdb 表实现 vxlan
目的：通过手动维护，就不需要通过多播方式获取对方 flannel.1 vxlan 网卡的 mac，和对方 node ip。

> https://xiaohanliang.gitbook.io/xiaohanliang/v/os/network-devices/vxlan-dfb#

* arp 表: 负责回答 mac 地址，flannel.1 ip-> flannel.1 mac，如果知道对方 flannel.1 的 ip 地址，则需要知道对方 flannel.1 的 mac 地址
* fdb 表：负责回答 ip 地址，flannel.1 mac-> node ip，知道对方 flannel.1 的 mac 地址，需要知道对方的宿主机 ip 地址

结论：所以，像 flannel.1 vxlan 网卡，node1 上会创建一条 arp 条目，一个 fdb 条目。以下是 vxlan 一个包的查表流程:

```shell
# (1) 根据路由，pod1 ping pod2，目的 ip 是 10.244.1.0，走本地 flannel.1
ip route
10.244.1.0/24 via 10.244.1.0 dev flannel.1 onlink

# (2) 10.244.1.0/72:ff:29:6f:e7:98 是 node2 上的 flannel.1 的 ip 和 mac，根据这个记录查到 node2 flannel.1 的 mac 地址
ip neigh show dev flannel.1
10.244.1.0 lladdr 72:ff:29:6f:e7:98 PERMANENT

# (3) ff:29:6f:e7:98 是 node1 上的 flannel.1 的 mac，根据这个记录查到 node2 的 ip 192.168.49.3
bridge fdb show dev flannel.1
72:ff:29:6f:e7:98 dst 192.168.49.3 self permanent
```


