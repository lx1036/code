



# Kubernetes学习笔记之Pod DNS源码解析

## Overview
本文章基于k8s release-1.17分支代码，代码位于 `pkg/kubelet/netwrok/dns` 目录，代码：**[dns.go](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/network/dns/dns.go)** 。













## 参考文献

**[官网：Pod 与 Service 的 DNS](https://kubernetes.io/zh/docs/concepts/services-networking/dns-pod-service/)**

**[Kubernetes DNS 高阶指南](https://juejin.cn/post/6844903665879220231)**

**[kubelet cli reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)**

**[通过配置文件设置 Kubelet 参数](https://kubernetes.io/zh/docs/tasks/administer-cluster/kubelet-config-file/)**

**[华为云Kubernetes集群内置DNS配置说明](https://support.huaweicloud.com/intl/zh-cn/usermanual-cce/cce_01_0133.html)**
