



# Kubernetes学习笔记之Calico CNI Plugin源码解析


## Overview
之前在 **[Kubernetes学习笔记之kube-proxy service实现原理](https://segmentfault.com/a/1190000038801963)** 学习到calico会在
worker节点上为pod创建路由route和虚拟网卡virtual interface，并为pod分配pod ip，以及为worker节点分配pod cidr网段。

我们生产k8s网络插件使用calico cni，在安装时会安装两个插件：calico和calico-ipam，官网安装文档 **[Install the plugin](https://docs.projectcalico.org/getting-started/kubernetes/hardway/install-cni-plugin#install-the-plugin)** 也说到了这一点，
而这两个插件代码在 **[calico.go](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/cmd/calico/calico.go)** ，代码会编译出两个二进制文件：calico和calico-ipam。
calico插件主要用来创建route和virtual interface，而calico-ipam插件主要用来分配pod ip和为worker节点分配pod cidr。

重要问题是，calico是如何做到的？


## Sandbox container
kubelet进程在开始启动时，会调用容器运行时的 **[SyncPod](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/kubelet.go#L1692)** 来创建pod内相关容器，
主要做了几件事情 **[L657-L856](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/kuberuntime/kuberuntime_manager.go#L657-L856)** ：

* 创建sandbox container，这里会调用cni插件创建network等步骤，同时考虑了边界条件，创建失败会kill sandbox container等等
* 创建ephemeral containers、 init containers和普通的containers。

这里只关注创建sandbox container过程，只有这一步会创建pod network，这个sandbox container创建好后，其余container都会和其共享同一个network namespace，
所以一个pod内各个容器看到的网络栈是同一个，ip地址都是相同的，通过pod来区分各个容器。
具体创建过程，会调用容器运行时服务创建容器，这里会先准备好pod的相关配置数据，创建network namespace时也需要这些配置数据 **[L36-L138](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/kuberuntime/kuberuntime_sandbox.go#L36-L138)** ：

```go

func (m *kubeGenericRuntimeManager) createPodSandbox(pod *v1.Pod, attempt uint32) (string, string, error) {
	// 生成pod相关配置数据
	podSandboxConfig, err := m.generatePodSandboxConfig(pod, attempt)
	// ...

	// 这里会在宿主机上创建pod logs目录，在/var/log/pods/{namespace_{pod_name}_{uid}目录下
	err = m.osInterface.MkdirAll(podSandboxConfig.LogDirectory, 0755)
	// ...

	podSandBoxID, err := m.runtimeService.RunPodSandbox(podSandboxConfig, runtimeHandler)
	// ...

	return podSandBoxID, "", nil
}

```




## K8s CNI



## calico plugin源码解析








## calico ipam plugin源码解析



## 总结







## 参考文献







