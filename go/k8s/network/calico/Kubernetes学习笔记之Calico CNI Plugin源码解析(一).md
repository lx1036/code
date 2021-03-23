

# Kubernetes学习笔记之Calico CNI Plugin源码解析(一)


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
* 创建ephemeral containers、init containers和普通的containers。

这里只关注创建sandbox container过程，只有这一步会创建pod network，这个sandbox container创建好后，其余container都会和其共享同一个network namespace，
所以一个pod内各个容器看到的网络协议栈是同一个，ip地址都是相同的，通过port来区分各个容器。
具体创建过程，会调用容器运行时服务创建容器，这里会先准备好pod的相关配置数据，创建network namespace时也需要这些配置数据 **[L36-L138](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/kuberuntime/kuberuntime_sandbox.go#L36-L138)** ：

```go

func (m *kubeGenericRuntimeManager) createPodSandbox(pod *v1.Pod, attempt uint32) (string, string, error) {
	// 生成pod相关配置数据
	podSandboxConfig, err := m.generatePodSandboxConfig(pod, attempt)
	// ...

	// 这里会在宿主机上创建pod logs目录，在/var/log/pods/{namespace}_{pod_name}_{uid}目录下
	err = m.osInterface.MkdirAll(podSandboxConfig.LogDirectory, 0755)
	// ...

	// 调用容器运行时创建sandbox container，我们生产k8s这里是docker创建
	podSandBoxID, err := m.runtimeService.RunPodSandbox(podSandboxConfig, runtimeHandler)
	// ...

	return podSandBoxID, "", nil
}

```

k8s使用cri(container runtime interface)来抽象出标准接口，目前docker还不支持cri接口，所以kubelet做了个适配模块dockershim，代码在 `pkg/kubelet/dockershim` 。
上面代码中的runtimeService对象就是dockerService对象，所以可以看下 `dockerService.RunPodSandbox()` 代码实现 **[L76-L197](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/dockershim/docker_sandbox.go#L76-L197)** ：

```go

// 创建sandbox container，以及为该container创建network
func (ds *dockerService) RunPodSandbox(ctx context.Context, r *runtimeapi.RunPodSandboxRequest) (*runtimeapi.RunPodSandboxResponse, error) {
	config := r.GetConfig()

	// Step 1: Pull the image for the sandbox.
	// 1. 拉取镜像
	image := defaultSandboxImage
	podSandboxImage := ds.podSandboxImage
	if len(podSandboxImage) != 0 {
		image = podSandboxImage
	}

	if err := ensureSandboxImageExists(ds.client, image); err != nil {
		return nil, err
	}

	// Step 2: Create the sandbox container.
	// 2. 创建sandbox container
	createResp, err := ds.client.CreateContainer(*createConfig)
	// ...
	resp := &runtimeapi.RunPodSandboxResponse{PodSandboxId: createResp.ID}
	ds.setNetworkReady(createResp.ID, false)

	// Step 3: Create Sandbox Checkpoint.
	// 3. 创建checkpoint
	if err = ds.checkpointManager.CreateCheckpoint(createResp.ID, constructPodSandboxCheckpoint(config)); err != nil {
		return nil, err
	}

	// Step 4: Start the sandbox container.
	// Assume kubelet's garbage collector would remove the sandbox later, if
	// startContainer failed.
	// 4. 启动容器
	err = ds.client.StartContainer(createResp.ID)
	// ...

	// Step 5: Setup networking for the sandbox.
	// All pod networking is setup by a CNI plugin discovered at startup time.
	// This plugin assigns the pod ip, sets up routes inside the sandbox,
	// creates interfaces etc. In theory, its jurisdiction ends with pod
	// sandbox networking, but it might insert iptables rules or open ports
	// on the host as well, to satisfy parts of the pod spec that aren't
	// recognized by the CNI standard yet.
	
	// 5. 这一步为sandbox container创建网络，主要是调用calico cni插件创建路由和虚拟网卡，以及为pod分配pod ip，为该宿主机划分pod网段
	cID := kubecontainer.BuildContainerID(runtimeName, createResp.ID)
	networkOptions := make(map[string]string)
	if dnsConfig := config.GetDnsConfig(); dnsConfig != nil {
		// Build DNS options.
		dnsOption, err := json.Marshal(dnsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dns config for pod %q: %v", config.Metadata.Name, err)
		}
		networkOptions["dns"] = string(dnsOption)
	}
	// 这一步调用网络插件来setup sandbox pod
	// 由于我们网络插件都是cni(container network interface)，所以代码在 pkg/kubelet/dockershim/network/cni/cni.go
	err = ds.network.SetUpPod(config.GetMetadata().Namespace, config.GetMetadata().Name, cID, config.Annotations, networkOptions)
	// ...

	return resp, nil
}
```


由于我们网络插件都是cni(container network interface)，代码 `ds.network.SetUpPod` 继续追下去发现实际调用的是 `cniNetworkPlugin.SetUpPod()`，代码在 **[pkg/kubelet/dockershim/network/cni/cni.go](https://github.com/kubernetes/kubernetes/blob/release-1.17/pkg/kubelet/dockershim/network/cni/cni.go#L300-L321)** ：

```go
func (plugin *cniNetworkPlugin) SetUpPod(namespace string, name string, id kubecontainer.ContainerID, annotations, options map[string]string) error {
	// ...
	netnsPath, err := plugin.host.GetNetNS(id.ID)
	// ...
	// Windows doesn't have loNetwork. It comes only with Linux
	if plugin.loNetwork != nil {
        // 添加loopback
		if _, err = plugin.addToNetwork(cniTimeoutCtx, plugin.loNetwork, name, namespace, id, netnsPath, annotations, options); err != nil {
			return err
		}
	}
    // 调用网络插件创建网络相关资源
	_, err = plugin.addToNetwork(cniTimeoutCtx, plugin.getDefaultNetwork(), name, namespace, id, netnsPath, annotations, options)
	return err
}

func (plugin *cniNetworkPlugin) addToNetwork(ctx context.Context, network *cniNetwork, podName string, podNamespace string, podSandboxID kubecontainer.ContainerID, podNetnsPath string, annotations, options map[string]string) (cnitypes.Result, error) {
	// 这一步准备网络插件所需相关参数，这些参数最后会被calico插件使用
    rt, err := plugin.buildCNIRuntimeConf(podName, podNamespace, podSandboxID, podNetnsPath, annotations, options)
    // ...
    // 这里会调用调用cni标准库里的AddNetworkList函数，最后会调用calico二进制命令
    res, err := cniNet.AddNetworkList(ctx, netConf, rt)
    // ...
    return res, nil
}
// 这些参数主要包括container id，pod等相关参数
func (plugin *cniNetworkPlugin) buildCNIRuntimeConf(podName string, podNs string, podSandboxID kubecontainer.ContainerID, podNetnsPath string, annotations, options map[string]string) (*libcni.RuntimeConf, error) {
    rt := &libcni.RuntimeConf{
        ContainerID: podSandboxID.ID,
        NetNS:       podNetnsPath,
        IfName:      network.DefaultInterfaceName,
        CacheDir:    plugin.cacheDir,
        Args: [][2]string{
            {"IgnoreUnknown", "1"},
            {"K8S_POD_NAMESPACE", podNs},
            {"K8S_POD_NAME", podName},
            {"K8S_POD_INFRA_CONTAINER_ID", podSandboxID.ID},
        },
    }
    
    // port mappings相关参数
    // ...
    
    // dns 相关参数
    // ...
    
    return rt, nil
}
```

`addToNetwork()` 函数会调用cni标准库里的 **[AddNetworkList](https://github.com/containernetworking/cni/blob/master/libcni/api.go#L400-L440)** 函数。CNI是容器网络标准接口Container Network Interface，
这个代码仓库提供了CNI标准接口的相关实现，所有K8s网络插件都必须实现该CNI代码仓库中的接口，K8s网络插件如何实现规范可见 **[SPEC.md](https://github.com/containernetworking/cni/blob/master/SPEC.md)** ，我们也可实现遵循该标准规范实现一个简单的网络插件。
所以kubelet、cni和calico的三者关系就是：kubelet调用cni标准规范代码包，cni调用calico插件二进制文件。cni代码包中的AddNetworkList相关代码如下 **[AddNetworkList](https://github.com/containernetworking/cni/blob/master/libcni/api.go#L400-L440)**：

```go

func (c *CNIConfig) addNetwork(ctx context.Context, name, cniVersion string, net *NetworkConfig, prevResult types.Result, rt *RuntimeConf) (types.Result, error) {
	c.ensureExec()
	pluginPath, err := c.exec.FindInPath(net.Network.Type, c.Path)
	// ...

	// pluginPath就是calico二进制文件路径，这里其实就是调用 calico ADD命令，并传递相关参数，参数也是上文描述的已经准备好了的
	// 参数传递也是写入了环境变量，calico二进制文件可以从环境变量里取值
	return invoke.ExecPluginWithResult(ctx, pluginPath, newConf.Bytes, c.args("ADD", rt), c.exec)
}

// AddNetworkList executes a sequence of plugins with the ADD command
func (c *CNIConfig) AddNetworkList(ctx context.Context, list *NetworkConfigList, rt *RuntimeConf) (types.Result, error) {
	// ...
	for _, net := range list.Plugins {
		result, err = c.addNetwork(ctx, list.Name, list.CNIVersion, net, result, rt)
		// ...
	}
    // ...

	return result, nil
}

```

以上pluginPath就是calico二进制文件路径，这里calico二进制文件路径参数是在启动kubelet时通过参数 `--cni-bin-dir` 传进来的，可见官网 **[kubelet command-line-tools-reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)** ，并且启动参数 `--cni-conf-dir` 包含cni配置文件路径，该路径包含cni配置文件内容类似如下：

```json

{
  "name": "k8s-pod-network",
  "cniVersion": "0.3.1",
  "plugins": [
    {
      "type": "calico",
      "log_level": "debug",
      "log_file_path": "/var/log/calico/cni/cni.log",
      "datastore_type": "kubernetes",
      "nodename": "minikube",
      "mtu": 1440,
      "ipam": {
          "type": "calico-ipam"
      },
      "policy": {
          "type": "k8s"
      },
      "kubernetes": {
          "kubeconfig": "/etc/cni/net.d/calico-kubeconfig"
      }
    },
    {
      "type": "portmap",
      "snat": true,
      "capabilities": {"portMappings": true}
    },
    {
      "type": "bandwidth",
      "capabilities": {"bandwidth": true}
    }
  ]
}

```

cni相关代码是个标准骨架，核心还是需要调用第三方网络插件来实现为sandbox创建网络资源。cni也提供了一些示例plugins，代码仓库见 **[containernetworking/plugins](https://github.com/containernetworking/plugins)** ，
并配有文档说明见 **[plugins docs](https://www.cni.dev/plugins/)** ，比如可以参考学习官网提供的 **[static IP address management plugin](https://www.cni.dev/plugins/ipam/static/)** 。

## 总结
总之，kubelet在创建sandbox container时候，会先调用cni插件命令，如 `calico ADD` 命令并通过环境变量传递相关命令参数，来给sandbox container创建network相关资源对象，比如calico会创建
route和virtual interface，以及为pod分配ip地址，和从集群网段cluster cidr中为当前worker节点分配pod cidr网段，并且会把这些数据写入到calico datastore数据库里。

所以，关键问题，还是得看calico插件代码是如何做的。


## 参考文献
**[Use a specific IP address with a pod](https://docs.projectcalico.org/networking/use-specific-ip)**
**[Calico IPAM源码解析](https://mp.weixin.qq.com/s/lyfeZh6VWWjXuLY8fl3ciw)**
**[calico,CNI的一种实现](https://www.yuque.com/baxiaoshi/tyado3/lvfa0b)**
**[containernetworking/cni](https://github.com/containernetworking/cni)**
**[projectcalico/cni-plugin](https://github.com/projectcalico/cni-plugin)**

