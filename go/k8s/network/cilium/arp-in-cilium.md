
# ARP in Cilium
```shell
# 抓包arp协议
tcpdump -i eth0 -nnee arp and host 20.206.230.25
tcpdump -i eth0 -nnee arp
```
解决的问题：


## Neighbor Discovery
https://docs.cilium.io/en/stable/gettingstarted/kubeproxy-free/#neighbor-discovery
https://isovalent.com/blog/post/2021-12-release-111#managed-ipv4-ipv6-discovery

**[Add a mechanism in neighbor management to allow refreshing ARP cache](https://github.com/cilium/cilium/issues/14322)** :
**[daemon, node: refresh neighbor by sending arping periodically](https://github.com/cilium/cilium/pull/14498)**



