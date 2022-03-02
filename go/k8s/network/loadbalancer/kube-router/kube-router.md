

# Kube-Router
Kube-Router CNI: https://github.com/cloudnativelabs/kube-router https://kube-router.io/

## Features
(1) IPVS/LVS Service Proxy
(2) Pod/ServiceIP Networking with BGP: gobgp BGP 包来宣告 pod cidr 和 ServiceIP/ExternalIP service
(3) Network Policy: ipsets with iptables 来实现的

### IPVS/LVS Service Proxy
kube-router 使用 IPVS 直接做 NodePort/ClusterIP service，这样岂不是不需要 kube-proxy？
@see demo: https://asciinema.org/a/120312 


### Pod/ServiceIP Networking with BGP
kube-router 使用 gobgp 包来实现 BGP 宣告路由，并且支持宣告 ClusterIP/ExternalIP 给交换机。这个功能很重要!!!
IPAM: kube-router 没有自己的 IPAM 来管理各个 node 的 pod subnet，而是使用了 kube-controller-manager IPAM 分配给各个 node 的 pod subnet。这个很重要!!!

(1) 获取 node 上的 pod cidr
```shell
# 获取每个 node 的 pod cidr，这个是 kube-controller-manager [IPAM](https://github.com/kubernetes/kubernetes/blob/v1.22.0/pkg/controller/nodeipam/node_ipam_controller.go) 管理的
kubectl get nodes -o json | jq '.items[] | .spec'
#{
#  "podCIDR": "10.216.136.0/24",
#  "podCIDRs": [
#    "10.216.136.0/24"
#  ]
#}
#{
#  "podCIDR": "10.216.137.0/24",
#  "podCIDRs": [
#    "10.216.137.0/24"
#  ]
#}
#{
#  "podCIDR": "10.216.139.0/24",
#  "podCIDRs": [
#    "10.216.139.0/24"
#  ]
#}

```


### Network Policy
kube-router 使用 iptables 来实现 NetworkPolicy。




## Troubleshot
(1) kube-router 有没有自己的 IPAM 来管理 service ip from service cidr 的分配？




# 笔记

(2) bridge-nf
https://feisky.gitbooks.io/sdn/content/linux/params.html#bridge-nf
bridge-nf使得netfilter可以对Linux网桥上的IPv4/ARP/IPv6包过滤。比如，设置net.bridge.bridge-nf-call-iptables＝1后，
二层的网桥在转发包时也会被iptables的FORWARD规则所过滤，这样有时会出现L3层的iptables rules去过滤L2的帧的问题

net.bridge.bridge-nf-call-iptables(/proc/sys/net/bridge/bridge-nf-call-iptables)：是否在iptables链中过滤IPv4包


(3) linux policy based route 策略路由
linux 策略路由: https://linuxgeeks.github.io/2017/03/17/170119-Linux%E7%9A%84%E7%AD%96%E7%95%A5%E8%B7%AF%E7%94%B1/
linux 支持路由可以有多个表，每个表包含自己的路由，同时可以添加路由策略。
