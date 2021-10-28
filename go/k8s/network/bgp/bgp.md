
## BGP
https://github.com/osrg/gobgp
**[bgp route server](https://github.com/osrg/gobgp/blob/master/docs/sources/route-server.md)**

使用两台服务器测试：
```toml
# 100.208.40.78 上执行：./gobgpd -f ./gobgpd.conf
[global.config]
  as = 64512
  router-id = "100.208.40.78"
  port = 1790
  local-address-list = ["0.0.0.0"]

[[neighbors]]
  [neighbors.config]
    neighbor-address = "100.208.40.96"
    peer-as = 65001
  [neighbors.transport.config]
    remote-port = 1790
```

```toml
# 100.208.40.96 上执行：./gobgpd -f ./gobgpd.conf
[global.config]
  as = 65001
  router-id = "100.208.40.96"
  port = 1790
  local-address-list = ["0.0.0.0"]

[[neighbors]]
  [neighbors.config]
    neighbor-address = "100.208.40.78"
    peer-as = 64512
  [neighbors.transport.config]
    remote-port = 1790
```


```shell
# 添加/查看路由
./gobgp global rib add -a ipv4 100.0.0.0/24 nexthop 20.20.20.20
./gobgp global rib
#   Network              Next Hop             AS_PATH              Age        Attrs
#*> 100.0.0.0/24          20.20.20.20                               00:00:36   [{Origin: ?}]

./gobgp neighbor 100.208.40.96 adj-in
./gobgp neighbor 100.208.40.96 adj-out
#ID  Network              Next Hop             AS_PATH              Attrs
#1   100.0.0.0/24          20.20.20.20          64512                [{Origin: ?}]


./gobgp neighbor
./gobgp neighbor 100.208.40.96
#BGP neighbor is 100.208.40.96, remote AS 65001
#  BGP version 4, remote router ID 1.1.1.1
#  BGP state = ESTABLISHED, up for 00:00:16
#  BGP OutQ = 0, Flops = 0
#  Hold time is 90, keepalive interval is 30 seconds
#  Configured hold time is 90, keepalive interval is 30 seconds
#
#  Neighbor capabilities:
#    multiprotocol:
#        ipv4-unicast:	advertised and received
#    route-refresh:	advertised and received
#    extended-nexthop:	advertised and received
#        Local:  nlri: ipv4-unicast, nexthop: ipv6
#        Remote: nlri: ipv4-unicast, nexthop: ipv6
#    4-octet-as:	advertised and received
#    fqdn:	advertised and received
#      Local:
#         name: docker07.example.net, domain:
#      Remote:
#         name: docker08.example.net, domain:
#  Message statistics:
#                         Sent       Rcvd
#    Opens:                  1          1
#    Notifications:          0          0
#    Updates:                0          0
#    Keepalives:             1          1
#    Route Refresh:          0          0
#    Discarded:              0          0
#    Total:                  2          2
#  Route statistics:
#    Advertised:             0
#    Received:               0
#    Accepted:               0
```

## kube-ovn 支持 bgp(使用 gobgp 包)
https://github.com/kubeovn/kube-ovn/wiki/BGP-%E6%94%AF%E6%8C%81
