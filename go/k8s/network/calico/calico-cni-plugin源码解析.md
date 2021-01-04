



# Kubernetes学习笔记之Calico CNI Plugin源码解析


## Overview
之前在 **[Kubernetes学习笔记之kube-proxy service实现原理](https://segmentfault.com/a/1190000038801963)** 学习到calico会在
worker节点上为pod创建路由route和虚拟网卡virtual interface，并为pod分配pod ip，以及为worker节点分配pod cidr网段。

我们生产k8s网络插件使用calico cni，在安装时会安装两个插件：calico和calico-ipam，官网安装文档 **[Install the plugin](https://docs.projectcalico.org/getting-started/kubernetes/hardway/install-cni-plugin#install-the-plugin)** 也说到了这一点，
而这两个插件代码在 **[calico.go](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/cmd/calico/calico.go)** ，代码会编译出两个二进制文件：calico和calico-ipam。
calico插件主要用来创建route和virtual interface，而calico-ipam插件主要用来分配pod ip和为worker节点分配pod cidr。

重要问题是，calico是如何做到的？


## K8s CNI



## calico plugin源码解析





## calico ipam plugin源码解析



## 总结







## 参考文献







