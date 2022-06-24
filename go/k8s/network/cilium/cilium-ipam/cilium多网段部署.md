
# Cilium 多网段 cilium-ipam operator 部署

## 背景
calico 可以支持一个 K8s 集群内配置多个 ippool, 并且可以根据 nodeSelector 选择对应的 ippool，比如 https://projectcalico.docs.tigera.io/getting-started/kubernetes/hardway/configure-ip-pools 。

但是，cilium 目前不支持根据 nodeSelector 来选择对应的 ippool, 且 cilium 也没有 ippool 这样的 crd，不得不说 cilium ipam(IP Address Management) 这块还比较差。

但是，cilium 可以支持没有 nodeSelector 选择器的多个 ippool，cilium 网段的传参是通过字符串数组，即一个集群网段耗尽后，会依次选择下一个网段，见 https://github.com/cilium/cilium/blob/v1.11.5/operator/flags.go#L182-L186 。所以这块可以解决这样问题：当选择一个集群网段后，随着不断加入机器耗尽集群网段，可以在该参数上 `--cluster-pool-ipv4-cidr` 上继续添加一个网段。 

总之，cilium ipam 只能支持没有 nodeSelector 选择器的多个 ippool，而不支持根据 nodeSelector 来为 node 分配指定的 ippool。所以，需要开发一个 cilium-ipam operator 来解决这个问题。

## cilium-ipam operator 原理
cilium ipam 包含几种模式，见 https://github.com/cilium/cilium/blob/v1.11.5/pkg/ipam/option/option.go#L6-L28 ，对于我们有用的是 kubernetes、crd 和 cluster-pool 三种模式。

目前生产 K8s 使用的 cilium 的 IPAM 使用的 是 cluster-pool 模式：为每一个 node 分配 pod cidr 交给部署的 cilium-operator 来做。但是，根据上文所述，这种模式不支持根据 nodeSelector 来选择对应的 ippool。
所以，我们这里选择 kubernetes 模式，为每一个 node 分配 pod cidr 的问题交给 K8s 去处理，然后 cilium 去获取这个 pod cidr。

cilium 去获取这个 pod cidr 的步骤是：
1. 从 K8s Node .Spec.PodCIDRs 或 .Spec.PodCIDR 里取 pod cidr
2. 如果第一步没有，则从 K8s Node Annotations["io.cilium.network.ipv4-pod-cidr"] 里取值。这里就是我们 operator 的可以做的文章。

代码可以见：
https://github.com/cilium/cilium/blob/v1.11.5/pkg/k8s/init.go#L109-L111
https://github.com/cilium/cilium/blob/v1.11.5/pkg/k8s/node.go#L114-L155

根据上文所述，cilium-ipam operator 代码原理就是：首先需要先关闭 kube-controller-manager `--allocate-node-cidrs=false` 给 K8s Node 分配 PodCIDR，让 cilium 从 annotation 里取值。
然后定义集群多个带有 nodeSelector 的 IPPool crd 对象，cilium-ipam operator 根据 node label 来选择对应的某一个 IPPool，然后从该 IPPool 中按照 IPPool.blockSize 来为该 K8s Node 分配
对应的 pod cidr 如 10.20.30.0/24，并给 K8s Node 打上 "io.cilium.network.ipv4-pod-cidr=10.20.30.0/24" annotation。最后，cilium 会自动读取这个 annotation 作为该 node 的 pod cidr。

所以，operator 代码整体上比较简单。

## 部署工作

(1) 需要确保 cilium 是最新版本，目前选择 cilium 1.11.5 版本，配置参数需要稍许修改，基本变化不大。configmap.yaml 会放于代码仓库 。然后部署 cilium 为尽可能新的版本。


(2) 关闭 kube-controller-manager `--allocate-node-cidrs=false` 。但是目前已经部署的 K8s 已经开启，K8s Node .Spec.PodCIDR 已经被分配值了，所以除了改变 `--allocate-node-cidrs=false` 参数外，
还需要重新生成 K8s Node：每一台都要人工删除 K8s Node 对象，然后重启该 Node 上的 kubelet 使其生成 .Spec.PodCIDR 没有值的 K8s Node 对象(该方法废弃，直接 patch 掉 Node .Spec.PodCIDR 就行)。


(3)最后一步就简单了，用 helm chart 部署 cilium-ipam operator。首次部署时，直接在代码仓库下 https://v.src.corp.qihoo.net/opsdev/loadbalancer 执行 `make cilim-ipam-install` 命令就行；
后续升级 operator 时，只需要执行 `make cilim-ipam-upgrade` 。代码仓库会做好 helm chart，以后只需简单的 make 命令就自动化部署。

所以，部署 cilium-ipam operator 很简单，但是 *尤其由于历史原因*，第二步的准备工作比较麻烦，尤其对于已经有大量业务的生产 K8s 来说，基本做不到平滑，只能是暂停整个 K8s 业务流量离线去搞。
所以，目前先从新生产 K8s 开始来搞，已有的 K8s 后续再结合需求来搞。
