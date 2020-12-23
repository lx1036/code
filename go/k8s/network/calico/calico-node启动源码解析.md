


# calico


# Kubernetes学习笔记之Calico Startup源码解析

## Overview
我们目前生产k8s calico使用ansible二进制部署在私有机房，没有使用官方的calico/node容器部署，并且因为没有使用network policy只部署了confd/bird进程服务，
没有部署felix。
采用BGP(Border Gateway Protocol)方式来部署网络，并且采用**[Peered with TOR (Top of Rack) routers](https://docs.projectcalico.org/networking/determine-best-networking#on-prem)**
方式部署，每一个worker node和其置顶交换机建立bgp peer配对，置顶交换机会继续后上层核心交换机建立bgp peer配对，这样可以保证pod ip在公司内网可以直接被访问。

> BGP: 主要是网络之间分发动态路由的一个协议，使用TCP协议传输数据。比如，交换机A下连着12台worker node，可以在每一台worker node上安装一个BGP Client，如Bird或GoBGP程序，
> 这样每一台worker node会把自己的路由分发给交换机A，交换机A会做路由聚合，以及继续向上一层核心交换机转发。

平时在维护k8s云平台时，有时发现一台worker节点上的所有pod的ip在集群外没法访问，经过排查发现是该worker节点有两张内网网卡eth0和eth1，eth0 IP地址和交换机建立BGP
连接，并获取其as number号，但是bird启动配置文件bird.cfg里使用的eth1网卡IP地址。 并且发现calico里的**[Node](https://docs.projectcalico.org/reference/resources/node)** 
数据的IP地址ipv4Address和 **[BGPPeer](https://docs.projectcalico.org/reference/resources/bgppeer)** 数据的交换机地址peerIP也对不上。

一番抓头挠腮后，找到根本原因是我们的ansible部署时，通过调用网络API与交换机建立bgp peer配对时，使用的是eth0地址，
并且通过ansible任务`calicoctl apply -f node_peer.yaml` 写入**[Node-specific BGP Peer](https://docs.projectcalico.org/reference/resources/bgppeer#node-specific-peer)**数据，
写入calico BGP Peer数据里的是eth0交换机地址。但是ansible任务跑到配置bird.cfg配置文件时，环境变量IP使用的是eth1 interface，
写入calico Node数据使用的是eth1网卡地址，然后被confd进程读取Node数据生成bird.cfg文件时，使用的就会是eth1网卡地址。

找到问题原因后，就愉快的解决了。

但是，又突然想知道，calico是怎么写入Node数据的？代码原来在calico启动代码 **[startup.go](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go)** 这里。
官方提供的calico/node容器




