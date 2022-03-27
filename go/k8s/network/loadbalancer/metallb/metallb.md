
# MetalLB
docs: https://github.com/metallb/metallb https://metallb.universe.tf/

## Features
(1) 监听 LoadBalancer Service，且自带 Service IPAM，且支持 Multi-IPPool。使用指定 IPPool，为每一个 LoadBalancer Service allocate service ip。
deployment 部署。Service IPAM 使用 K8s 源码自带的 IPAM，可以参考 ClusterIP Service IPAM。
(2) 根据 service ExternalTrafficPolicy 来 BGP 宣告 service ip。daemonset 部署。
