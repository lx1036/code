


# CNI
代码仓库：github.com/containernetworking/cni
plugins: https://github.com/containernetworking/plugins
cni debug tool: `go install github.com/containernetworking/cni/cnitool`


# 设计 CNI
设计需要考虑的问题：
(1) 网络连通性问题：pod 和 pod, pod 和 service, pod 和 node, 以及 pod 和外部网络。
(2) cni binary(与 kubelet 交互) 和 ipam daemon 职责划分。
(3) 高效的 ipam 资源划分和使用。
(4) 不同的机器(公有云机器型号不一样，支持的弹性网卡数量也不一样)网络资源配额可能不一致，如何让调度感知。
(5) 异常处理，垃圾资源的回收。


## ENI 弹性网卡
弹性网卡概念: https://help.aliyun.com/document_detail/58496.html

## VSwitch 虚拟交换机
VSwitch: https://help.aliyun.com/document_detail/65387.htm



## 参考文献
**[使用 Go 从零开始实现 CNI](https://morven.life/posts/write_your_own_cni_with_golang/)**
**[Linux 数据包的接收与发送过程](https://morven.life/posts/networking-1-pkg-snd-rcv/)**
**[Linux 虚拟网络设备](https://morven.life/posts/networking-2-virtual-devices/)**
**[从 container 到 pod](https://morven.life/posts/from-container-to-pod/)**
**[容器网络(一)](https://morven.life/posts/networking-4-docker-sigle-host/)**
**[容器网络(二)](https://morven.life/posts/networking-5-docker-multi-hosts/)**
**[浅聊 Kubernetes 网络模型](https://morven.life/posts/networking-6-k8s-summary/)**
