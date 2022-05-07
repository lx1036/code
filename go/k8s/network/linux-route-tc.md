
# Linux Advanced Routing & Traffic Control 
https://lartc.org/howto/

总结：数据包经过网卡 xdp -> tc ingress -> netfilter -> tc egress 时，在经过 netfilter 时，在 PREROUTING chain 里做 DNAT，
POSTROUTING chain 里做 SNAT，然后开始路由转发。而路由数据库在各个路由表里，所以需要 ip rule list 查找各个路由策略根据一系列规则判断出选择
哪个路由表，缺省 main 路由表。然后根据选择的路由表的路由，决定下一跳。


```shell
ip rule add # 路由策略
ip route add # 路由

tc qdisc add # 下发 tc ingress/egress/clsact 流量整形规则，或者包转发给另一个网卡
tc filter add
```

## Routing

### route policy(使用 ip rule 命令操作路由策略数据库)
基于策略的路由比传统路由在功能上更强大，使用更灵活，它使网络管理员不仅能够根据目的地址，而且能够根据报文大小、应用或IP源地址等属性来选择转发路径

```shell
# SELECTOR := [ from PREFIX 数据包源地址] [ to PREFIX 数据包目的地址] [ tos TOS 服务类型][ dev STRING 物理接口] [ pref NUMBER ] [fwmark MARK iptables 标签]
# ACTION := [ table TABLE_ID 指定所使用的路由表] [ nat ADDRESS 网络地址转换][ prohibit 丢弃该表| reject 拒绝该包| unreachable 丢弃该包]

Usage: ip rule { add | del } SELECTOR ACTION
       ip rule { flush | save | restore }
       ip rule [ list [ SELECTOR ]]
SELECTOR := [ not ] [ from PREFIX ] [ to PREFIX ] [ tos TOS ] [ fwmark FWMARK[/MASK] ]
            [ iif STRING ] [ oif STRING ] [ pref NUMBER ] [ l3mdev ]
            [ uidrange NUMBER-NUMBER ]
ACTION := [ table TABLE_ID ]
          [ nat ADDRESS ]
          [ realms [SRCREALM/]DSTREALM ]
          [ goto NUMBER ]
          SUPPRESSOR
SUPPRESSOR := [ suppress_prefixlength NUMBER ]
              [ suppress_ifgroup DEVGROUP ]
TABLE_ID := [ local | main | default | NUMBER ]


ip rule add to 192.168.100.0/24 table kube-router dev eth0 prio 512 # 

```


问题：公司要求内网网段在 192.168.0.1-192.168.0.100 IP 使用电信网关 10.0.0.1 上网，其余的 IP 使用网通网关 20.0.0.1 上网？
```shell
ip route add default via 20.0.0.1 dev eth1

# mangle 的处理是优先于 nat 和 fiter 表的，所以在数据包到达之后先打上标记，之后再通过 ip rule 规则，对应的数据包使用相应的路由表进行路由，最后读取路由表信息，将数据包送出网关

# 先使用 iptables 给 192.168.0.1-192.168.0.100 IP 的包打上 mark
iptables -A PREROUTING -t mangle -i eth0 -s 192.168.0.1-192.168.0.100 -j MARK --set-mark 0x10000/0x10000 # PREROUTING chain 里做 DNAT，POSTROUTING chain 里做 SNAT
# 查询路由策略，然后走路由表的路由
echo "77 kube-router" >> "/etc/iproute2/rt_tables"
ip rule add fwmark 0x10000/0x10000 table kube-router # 凡是有 mark 0x10000/0x10000 的 packet 查询 kube-router 路由表
ip route add table kube-router default via 10.0.0.1 dev eth0
```


