

# OpenLB
https://porterlb.io/
https://github.com/kubesphere/openelb
https://github.com/metallb/metallb

OpenLB: 类似于 MetalLB(https://github.com/metallb/metallb)，使用 gobgp(https://github.com/osrg/gobgp) 库走 bgp 协议来宣告路由到交换机对端，
让得 k8s loadbalancer service ip 可以集群外访问。 见：https://porterlb.io/docs/concepts/bgp-mode/

## Features
(1) BGP LoadBalancer Service IP，且自带 Service IPAM，且支持 Multi-IPPool，且支持 CRD 配置更友好。
比 MetalLB 配置使用更友好的 CRD 模式。

(2) LoadBalancer Service IPAM Controller allocate IP 和 BGP Speaker 宣告 pod cidr、loadbalancer service ExternalIP。
