
# BGP
BGP使用 TCP 作为传输协议，使用端口 179 建立连接，允许基于策略的路由，可以使用路由策略在目标的多个路径之间选择，并控制路由信息的再分配。

BGP 基于路径矢量(Path Vector)的算法,即每个路由条目更新通过一个AS时, 将其穿越的AS号码记录下来, 通过避免路径属性中出现相同的AS号码来避免环路的策略。

概念名词：
* ROA: Route Origin Authorization
* BMP: BGP Monitoring Protocol, provides a convenient interface for obtaining route views
* VRF: virtual routing and forwarding(Configuring a VRF to Provide BGP VPN Services)
* FIB: Forwarding Information Based 每个路由器本身有一个路由信息数据库(RIB Routing Information Based)，然后会根据本地算法来构建转发数据库(FIB)
* CIDR: Classless Inter-Domain Routing 无类域间路由
* AS: Autonomous System 自治系统

## GoBGP
```shell
# mac本地安装
go get github.com/osrg/gobgp/cmd/gobgp@v2.32.0
go get github.com/osrg/gobgp/cmd/gobgpd@v2.32.0
```

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

# Route Server
Route Server: https://github.com/osrg/gobgp/blob/master/docs/sources/route-server.md
route server 就类似于交换机那一侧，可以使用 gobgp 来模拟路由器那一侧。
这里 openlb 使用 bird 来模拟路由器：https://github.com/kubesphere/openelb/blob/master/doc/zh/simulate_with_bird.md


## RIB
RIB： Routing Information Base，由三部分组成：
The Adj-RIBs-In: BGP RIB-In stores BGP routing information received from different peers. 
The stored information is used as an input to BGP decision process. In other words this is the information received from
peers before applying any attribute modifications or route filtering to them.

The Local RIB: The local routing information base stores the resulted information from processing the RIBs-In database’s information. 
These are the routes that are used locally after applying BGP policies and decision process.

The Adj-RIBs-out: This one stores the routing information that was selected by the local BGP router to advertise to its peers through BGP update messages. Do not forget;  BGP only advertises best routes if they are allowed by local outbound policies.


## kube-ovn 支持 bgp(使用 gobgp 包)
https://github.com/kubeovn/kube-ovn/wiki/BGP-%E6%94%AF%E6%8C%81


# 参考文献
**[bgp](https://datatracker.ietf.org/doc/html/rfc4271)**

**[bgp-route-server](https://datatracker.ietf.org/doc/html/rfc7947)**

**[bgp-route-reflection](https://datatracker.ietf.org/doc/html/rfc4456)**

**[在FB闯祸的BGP协议简介(BGP 最佳文档!!!)](https://mp.weixin.qq.com/s/XXi03wNMTjejZJpN6YPQ9g)**
