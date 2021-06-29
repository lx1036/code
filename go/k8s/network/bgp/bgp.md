
## BGP
**[bgp route server](https://github.com/osrg/gobgp/blob/master/docs/sources/route-server.md)**


```toml
# ./gobgpd -f ./gobgpd.conf
# 作为 route server，类似交换机那端

[global.config]
  as = 64512
  router-id = "local-ip"
  port = 179
  local-address-list = ["local-ip"]
[[neighbors]]
  [neighbors.config]
    neighbor-address = "remote-ip"
    peer-as = 65001
  [neighbors.transport.config]
    passive-mode = true
  [neighbors.route-server.config]
    route-server-client = true

```

```toml
# ./gobgpd -f ./gobgpd.conf

[global.config]
  as = 65001
  router-id = "local-ip"
  port = 179
  local-address-list = ["local-ip"]
[[neighbors]]
  [neighbors.config]
    neighbor-address = "remote-ip"
    peer-as = 64512

```
