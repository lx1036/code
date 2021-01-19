



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



# Kubernetes学习笔记之Calico CNI Plugin源码解析(二)


## Overview
calico插件代码仓库在 **[projectcalico/cni-plugin](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/cmd/calico/calico.go#L29-L45)** ，并且会编译两个二进制文件：calico和calico-ipam，其中
calico会为sandbox container创建route和虚拟网卡virtual interface，以及veth pair等网络资源，并且会把相关数据写入calico datastore数据库里；calico-ipam会为当前pod从当前节点的pod网段内分配ip地址，
当然当前节点还没有pod网段，会从集群网段cluster cidr中先分配出该节点的pod cidr，并把相关数据写入calico datastore数据库里，这里cluster cidr是用户自己定义的，已经提前写入了calico datastore，并且从cluster cidr中
划分的block size也是可以自定义的(新版本calico/node容器可以支持自定义，老版本calico不支持)，可以参考官网文档 **[change-block-size](https://docs.projectcalico.org/networking/change-block-size#concepts)** 。

接下来重点看下calico二进制插件具体是如何工作的，后续再看calico-ipam二进制插件如何分配ip地址的。

## calico plugin源码解析
calico插件是遵循cni标准接口，实现了 `ADD` 和 `DEL` 命令，这里重点看看 `ADD` 命令时如何实现的。calico首先会注册 `ADD` 和 `DEL` 命令，代码在 **[L614-L677](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/plugin/plugin.go#L614-L677)** ：

```go

func Main(version string) {
	// ...
	err := flagSet.Parse(os.Args[1:])
	// ...
	// 注册 `ADD` 和 `DEL` 命令
	skel.PluginMain(cmdAdd, nil, cmdDel,
		cniSpecVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1"),
		"Calico CNI plugin "+version)
}

```

`ADD` 命令里，主要做了三个逻辑：
* 查询calico datastore里有没有WorkloadEndpoint对象和当前的pod名字匹配，没有匹配，则会创建新的WorkloadEndpoint对象，
  该对象内主要保存该pod在host network namespace内的网卡名字和pod ip地址，以及container network namespace的网卡名字等等信息，对象示例如下。
* 创建一个veth pair，并把其中一个网卡置于宿主机端网络命名空间，另一个置于容器端网络命名空间。在container network namespace内创建网卡如eth0，并通过调用calico-ipam获得的IP地址赋值给该eth0网卡；
  在host network namespace内创建网卡，网卡名格式为 `"cali" + sha1(namespace.pod)[:11]` ，并设置MAC地址"ee:ee:ee:ee:ee:ee"。
* 在容器端和宿主机端创建路由。在容器端，设置默认网关为 `169.254.1.1` ，该网关地址代码写死的；在宿主机端，添加路由如 `10.217.120.85 dev calid0bda9976d5 scope link` ，
  其中 `10.217.120.85` 是pod ip地址，`calid0bda9976d5` 是该pod在宿主机端的网卡，也就是veth pair在宿主机这端的virtual ethernet interface虚拟网络设备。
  

一个WorkloadEndpoint对象示例如下，一个k8s pod对象对应着calico中的一个workloadendpoint对象，可以通过 `calicoctl get wep -o wide` 查看所有 workloadendpoint。
记得配置calico datastore为kubernetes的，为方便可以在 `~/.zshrc` 里配置环境变量：

```shell
# calico
export CALICO_DATASTORE_TYPE=kubernetes
export  CALICO_KUBECONFIG=~/.kube/config
```

```yaml

apiVersion: projectcalico.org/v3
kind: WorkloadEndpoint
metadata:
  creationTimestamp: "2021-01-09T08:38:56Z"
  generateName: nginx-demo-1-7f67f8bdd8-
  labels:
    app: nginx-demo-1
    pod-template-hash: 7f67f8bdd8
    projectcalico.org/namespace: default
    projectcalico.org/orchestrator: k8s
    projectcalico.org/serviceaccount: default
  name: minikube-k8s-nginx--demo--1--7f67f8bdd8--d5wsc-eth0
  namespace: default
  resourceVersion: "557760"
  uid: 85d1d33f-f55f-4f28-a89d-0a55394311db
spec:
  endpoint: eth0
  interfaceName: calife8e5922caa
  ipNetworks:
  - 10.217.120.84/32
  node: minikube
  orchestrator: k8s
  pod: nginx-demo-1-7f67f8bdd8-d5wsc
  profiles:
  - kns.default
  - ksa.default.default

```

根据以上三个主要逻辑，看下 **[cmdAdd](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/plugin/plugin.go#L105-L494)** 函数代码：

```go

func cmdAdd(args *skel.CmdArgs) (err error) {
    // ...
	// 从args.StdinData里加载配置数据，这些配置数据其实就是
	// `--cni-conf-dir` 传进来的文件内容，即cni配置参数，见第一篇文章
	// types.NetConf 结构体数据结构也对应着cni配置文件里的数据
	conf := types.NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}
	// 这里可以通过cni参数设置，把calico插件的日志落地到宿主机文件内
	// "log_level": "debug", "log_file_path": "/var/log/calico/cni/cni.log",
	utils.ConfigureLogging(conf)
	
	// ...
	
	// 可以在cni文件内设置MTU，即Max Transmit Unit最大传输单元，配置网卡时需要
	if mtu, err := utils.MTUFromFile("/var/lib/calico/mtu"); err != nil {
		return fmt.Errorf("failed to read MTU file: %s", err)
	} else if conf.MTU == 0 && mtu != 0 {
		conf.MTU = mtu
	}

	// 构造一个WEPIdentifiers对象，并赋值
	nodename := utils.DetermineNodename(conf)
	wepIDs, err := utils.GetIdentifiers(args, nodename)
	calicoClient, err := utils.CreateClient(conf)
	
	// 检查datastore是否已经ready了，可以 `calicoctl get clusterinformation default -o yaml` 查看
	ci, err := calicoClient.ClusterInformation().Get(ctx, "default", options.GetOptions{})
	if !*ci.Spec.DatastoreReady {
		return
	}

	// list出前缀为wepPrefix的workloadEndpoint，一个pod对应一个workloadEndpoint，如果数据库里能匹配出workloadEndpoint，就使用这个workloadEndpoint
	// 否则最后创建完pod network资源后，会往calico数据库里写一个workloadEndpoint
	wepPrefix, err := wepIDs.CalculateWorkloadEndpointName(true)
	endpoints, err := calicoClient.WorkloadEndpoints().List(ctx, options.ListOptions{Name: wepPrefix, Namespace: wepIDs.Namespace, Prefix: true})
	if err != nil {
		return
	}

	// 对于新建的pod，最后会在calico datastore里写一个对应的新的workloadendpoint对象
    var endpoint *api.WorkloadEndpoint

	// 这里因为我们是新建的pod，数据库里也不会有对应的workloadEndpoint对象，所以endpoints必然是nil的
	if len(endpoints.Items) > 0 {
		// ...
	}

	// 既然endpoint是nil，则填充WEPIdentifiers对象默认值，这里args.IfName是kubelet那边传过来的，就是容器端网卡名字，一般是eth0
	// 这里WEPName的格式为：{node_name}-k8s-{strings.replace(pod_name, "-", "--")}-{wepIDs.Endpoint}，比如上文
	// minikube-k8s-nginx--demo--1--7f67f8bdd8--d5wsc-eth0 WorkloadEndpoint对象
	if endpoint == nil {
		wepIDs.Endpoint = args.IfName
		wepIDs.WEPName, err = wepIDs.CalculateWorkloadEndpointName(false)
	}

	// Orchestrator是k8s
	if wepIDs.Orchestrator == api.OrchestratorKubernetes {
		// k8s.CmdAddK8s 函数里做以上三个逻辑工作
		if result, err = k8s.CmdAddK8s(ctx, args, conf, *wepIDs, calicoClient, endpoint); err != nil {
			return
		}
	} else {
		// ...
	}

	// 我们的配置文件里 policy.type 是k8s，可见上文配置文件
	if conf.Policy.PolicyType == "" {
		// ...
	}

	// Print result to stdout, in the format defined by the requested cniVersion.
	err = cnitypes.PrintResult(result, conf.CNIVersion)
	return
}

```

以上cmdAdd()函数基本结构符合cni标准里的函数结构，最后会把结果打印到stdout。看下 **[k8s.CmdAddK8s()](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/k8s/k8s.go#L48-L482)** 函数的主要逻辑：

```go
// 主要做三件事：
// 1. 往calico store里写个WorkloadEndpoint对象，和pod对应
// 2. 创建veth pair，一端在容器端，并赋值IP/MAC地址；一端在宿主机端，赋值MAC地址
// 3. 创建路由，容器端创建默认网关路由；宿主机端创建该pod ip/mac的路由
func CmdAddK8s(ctx context.Context, args *skel.CmdArgs, conf types.NetConf, epIDs utils.WEPIdentifiers, calicoClient calicoclient.Interface, endpoint *api.WorkloadEndpoint) (*current.Result, error) {
	// ...
	// 这里根据操作系统生成不同的数据平面data plane，这里是linuxDataplane对象
	d, err := dataplane.GetDataplane(conf, logger)
	// 创建k8s client
	client, err := NewK8sClient(conf, logger)
	
	// 我们的配置文件里 ipam.type=calico-ipam
	if conf.IPAM.Type == "host-local" {
		// ...
	}

	// ...
	
	// 这里会检查该pod和namespace的annotation: cni.projectcalico.org/ipv4pools
	// 我们没有设置，这里逻辑跳过
	if conf.Policy.PolicyType == "k8s" {
		annotNS, err := getK8sNSInfo(client, epIDs.Namespace)
        labels, annot, ports, profiles, generateName, err = getK8sPodInfo(client, epIDs.Pod, epIDs.Namespace)
		// ...
		if conf.IPAM.Type == "calico-ipam" {
			var v4pools, v6pools string
			// Sets  the Namespace annotation for IP pools as default
			v4pools = annotNS["cni.projectcalico.org/ipv4pools"]
			v6pools = annotNS["cni.projectcalico.org/ipv6pools"]
			// Gets the POD annotation for IP Pools and overwrites Namespace annotation if it exists
			v4poolpod := annot["cni.projectcalico.org/ipv4pools"]
			if len(v4poolpod) != 0 {
				v4pools = v4poolpod
			}
			// ...
		}
	}

	ipAddrsNoIpam := annot["cni.projectcalico.org/ipAddrsNoIpam"]
	ipAddrs := annot["cni.projectcalico.org/ipAddrs"]
	
	switch {
	// 主要走这个逻辑：调用calico-ipam插件分配一个IP地址
	case ipAddrs == "" && ipAddrsNoIpam == "":
		// 我们的pod没有设置annotation "cni.projectcalico.org/ipAddrsNoIpam"和"cni.projectcalico.org/ipAddrs"值
		// 这里调用calico-ipam插件获取pod ip值
		// 有关calico-ipam插件如何分配pod ip值，后续有空再学习下
		result, err = utils.AddIPAM(conf, args, logger)
		// ...
	case ipAddrs != "" && ipAddrsNoIpam != "":
		// Can't have both ipAddrs and ipAddrsNoIpam annotations at the same time.
		e := fmt.Errorf("can't have both annotations: 'ipAddrs' and 'ipAddrsNoIpam' in use at the same time")
		logger.Error(e)
		return nil, e
	case ipAddrsNoIpam != "":
		// ...
	case ipAddrs != "":
		// ...
	}
	
	// 开始创建WorkloadEndpoint对象，赋值相关参数
	endpoint.Name = epIDs.WEPName
	endpoint.Namespace = epIDs.Namespace
	endpoint.Labels = labels
	endpoint.GenerateName = generateName
	endpoint.Spec.Endpoint = epIDs.Endpoint
	endpoint.Spec.Node = epIDs.Node
	endpoint.Spec.Orchestrator = epIDs.Orchestrator
	endpoint.Spec.Pod = epIDs.Pod
	endpoint.Spec.Ports = ports
	endpoint.Spec.IPNetworks = []string{}
	if conf.Policy.PolicyType == "k8s" {
		endpoint.Spec.Profiles = profiles
	} else {
		endpoint.Spec.Profiles = []string{conf.Name}
	}

	// calico-ipam分配的ip地址值，写到endpoint.Spec.IPNetworks中
	if err = utils.PopulateEndpointNets(endpoint, result); err != nil {
		// ...
	}

	// 这里desiredVethName网卡名格式为：`"cali" + sha1(namespace.pod)[:11]` ，这个网卡为置于宿主机一端
	desiredVethName := k8sconversion.NewConverter().VethNameForWorkload(epIDs.Namespace, epIDs.Pod)
	
	// DoNetworking()函数很重要，该函数会创建veth pair和路由
	// 这里是调用linuxDataplane对象的DoNetworking()函数
	hostVethName, contVethMac, err := d.DoNetworking(
		ctx, calicoClient, args, result, desiredVethName, routes, endpoint, annot)
    
	// ...
	mac, err := net.ParseMAC(contVethMac)
	endpoint.Spec.MAC = mac.String()
	endpoint.Spec.InterfaceName = hostVethName
	endpoint.Spec.ContainerID = epIDs.ContainerID

	// ...

	// 创建或更新WorkloadEndpoint对象，至此到这里，会根据新建的一个pod对象，往calico datastore里写一个对应的workloadendpoint对象
	if _, err := utils.CreateOrUpdate(ctx, calicoClient, endpoint); err != nil {
		// ...
	}

	// Add the interface created above to the CNI result.
	result.Interfaces = append(result.Interfaces, &current.Interface{
		Name: endpoint.Spec.InterfaceName},
	)

	return result, nil
}

```

以上代码最后会创建个workloadendpoint对象，同时DoNetworking()函数很重要，这个函数里会创建路由和veth pair。
然后看下linuxDataplane对象的 **[DoNetworking()](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/dataplane/linux/dataplane_linux.go#L52-L352)** 函数，是如何创建veth pair和routes的。
这里主要调用了 `github.com/vishvananda/netlink` golang包来增删改查网卡和路由等操作，等同于执行 `ip link add/delete/set xxx` 等命令，
该golang包也是个很好用的包，被很多主要项目如k8s项目使用，在学习linux网络相关知识时可以利用这个包写一写相关demo，效率也高很多。这里看看calico如何使用netlink这个包来创建routes和veth pair的：

```go

func (d *linuxDataplane) DoNetworking(
	ctx context.Context,
	calicoClient calicoclient.Interface,
	args *skel.CmdArgs,
	result *current.Result,
	desiredVethName string,
	routes []*net.IPNet,
	endpoint *api.WorkloadEndpoint,
	annotations map[string]string,
) (hostVethName, contVethMAC string, err error) {
	// 这里desiredVethName网卡名格式为：`"cali" + sha1(namespace.pod)[:11]` ，这个网卡为置于宿主机一端
	hostVethName = desiredVethName
	// 容器这端网卡名一般为eth0
	contVethName := args.IfName

	err = ns.WithNetNSPath(args.Netns, func(hostNS ns.NetNS) error {
		veth := &netlink.Veth{
			LinkAttrs: netlink.LinkAttrs{
				Name: contVethName,
				MTU:  d.mtu,
			},
			PeerName: hostVethName,
		}
        // 创建veth peer，容器端网卡名是eth0，宿主机端网卡名是"cali" + sha1(namespace.pod)[:11]
        // 等于 ip link add xxx type veth peer name xxx 命令
        if err := netlink.LinkAdd(veth); err != nil {
		}
		hostVeth, err := netlink.LinkByName(hostVethName)
		if mac, err := net.ParseMAC("EE:EE:EE:EE:EE:EE"); err != nil {
		} else {
			// 设置宿主机端网卡的mac地址，为 ee:ee:ee:ee:ee:ee
			if err = netlink.LinkSetHardwareAddr(hostVeth, mac); err != nil {
				d.logger.Warnf("failed to Set MAC of %q: %v. Using kernel generated MAC.", hostVethName, err)
			}
		}

		// ...
		hasIPv4 = true

		// ip link set up起来宿主机端这边的网卡
		if err = netlink.LinkSetUp(hostVeth); err != nil {
		}
		// ip link set up起来容器端这边的网卡
		contVeth, err := netlink.LinkByName(contVethName)
		if err = netlink.LinkSetUp(contVeth); err != nil {
		}
		// Fetch the MAC from the container Veth. This is needed by Calico.
		contVethMAC = contVeth.Attrs().HardwareAddr.String()
		if hasIPv4 {
			// 容器端这边添加默认网关路由，如：
			// default via 169.254.1.1 dev eth0
			// 169.254.1.1 dev eth0 scope link
			gw := net.IPv4(169, 254, 1, 1)
			gwNet := &net.IPNet{IP: gw, Mask: net.CIDRMask(32, 32)}
			err := netlink.RouteAdd(
				&netlink.Route{
					LinkIndex: contVeth.Attrs().Index,
					Scope:     netlink.SCOPE_LINK,
					Dst:       gwNet,
				},
			)
		}

		// 把从calico-ipam插件分配来的pod ip地址赋值给容器端这边的网卡
		for _, addr := range result.IPs {
			if err = netlink.AddrAdd(contVeth, &netlink.Addr{IPNet: &addr.Address}); err != nil {
				return fmt.Errorf("failed to add IP addr to %q: %v", contVeth, err)
			}
		}
        // ...
		// 切换到宿主机端network namespace
		if err = netlink.LinkSetNsFd(hostVeth, int(hostNS.Fd())); err != nil {
			return fmt.Errorf("failed to move veth to host netns: %v", err)
		}

		return nil
	})

    // 设置veth pair宿主机端的网卡sysctls配置，设置这个网卡可以转发和arp_proxy
    err = d.configureSysctls(hostVethName, hasIPv4, hasIPv6)

	// ip link set up起来宿主机这端的veth pair的网卡
	hostVeth, err := netlink.LinkByName(hostVethName)
	if err = netlink.LinkSetUp(hostVeth); err != nil {
		return "", "", fmt.Errorf("failed to set %q up: %v", hostVethName, err)
	}

	// 配置宿主机这端的路由
	err = SetupRoutes(hostVeth, result)

	return hostVethName, contVethMAC, err
}

func SetupRoutes(hostVeth netlink.Link, result *current.Result) error {
	// 配置宿主机端这边的路由，凡是目的地址为pod ip 10.217.120.85，数据包进入calid0bda9976d5网卡，路由如：
	// 10.217.120.85 dev calid0bda9976d5 scope link
	for _, ipAddr := range result.IPs {
		route := netlink.Route{
			LinkIndex: hostVeth.Attrs().Index,
			Scope:     netlink.SCOPE_LINK,
			Dst:       &ipAddr.Address,
		}
		err := netlink.RouteAdd(&route)
        // ...
	}
	return nil
}

// 这里英文就不翻译解释了，英文备注说的更详细通透。

// configureSysctls configures necessary sysctls required for the host side of the veth pair for IPv4 and/or IPv6.
func (d *linuxDataplane) configureSysctls(hostVethName string, hasIPv4, hasIPv6 bool) error {
  var err error
  if hasIPv4 {
    // Normally, the kernel has a delay before responding to proxy ARP but we know
    // that's not needed in a Calico network so we disable it.
    if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/neigh/%s/proxy_delay", hostVethName), "0"); err != nil {
        return fmt.Errorf("failed to set net.ipv4.neigh.%s.proxy_delay=0: %s", hostVethName, err)
    }
    
    // Enable proxy ARP, this makes the host respond to all ARP requests with its own
    // MAC. We install explicit routes into the containers network
    // namespace and we use a link-local address for the gateway.  Turing on proxy ARP
    // means that we don't need to assign the link local address explicitly to each
    // host side of the veth, which is one fewer thing to maintain and one fewer
    // thing we may clash over.
    if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/proxy_arp", hostVethName), "1"); err != nil {
        return fmt.Errorf("failed to set net.ipv4.conf.%s.proxy_arp=1: %s", hostVethName, err)
    }
    
    // Enable IP forwarding of packets coming _from_ this interface.  For packets to
    // be forwarded in both directions we need this flag to be set on the fabric-facing
    // interface too (or for the global default to be set).
    if err = writeProcSys(fmt.Sprintf("/proc/sys/net/ipv4/conf/%s/forwarding", hostVethName), "1"); err != nil {
        return fmt.Errorf("failed to set net.ipv4.conf.%s.forwarding=1: %s", hostVethName, err)
    }
  }

  if hasIPv6 {
     // ...	
  }
  
  return nil
}
```



## 总结
至此，calico二进制插件就为一个sandbox container创建好了网络资源，即创建了一个veth pair，并分别为宿主机端和容器端网卡设置好对应MAC地址，以及为容器段配置好了IP地址，同时
还在容器端配置好了路由默认网关，以及宿主机端配置好路由，让目标地址是sandbox container ip的进入宿主机端veth pair网卡，同时还为宿主机端网卡配置arp proxy和packet forwarding功能，
最后，会根据这些网络数据生成一个workloadendpoint对象存入calico datastore里。

但是，还是缺少了一个关键逻辑，calico-ipam是如何分配IP地址的，后续有空在学习记录。


## 参考文献




# Kubernetes学习笔记之Calico CNI Plugin源码解析(三)

## Overview
从第二篇文章知道calico二进制插件会调用calico-ipam二进制插件，来为sandbox container分配一个IP地址，接下来重点看看 **[calico-ipam](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/ipamplugin/ipam_plugin.go)** 插件代码。


## calico ipam plugin源码解析
同样道理，calico-ipam插件也会注册cni的 `ADD` 和 `DEL` 命令，这里重点看看 `ADD` 命令都做了哪些工作 **[L115-L286)](https://github.com/projectcalico/cni-plugin/blob/release-v3.17/pkg/ipamplugin/ipam_plugin.go#L115-L286)**：

```go

func Main(version string) {
	// ...
	skel.PluginMain(cmdAdd, nil, cmdDel,
		cniSpecVersion.PluginSupports("0.1.0", "0.2.0", "0.3.0", "0.3.1"),
		"Calico CNI IPAM "+version)
}

type ipamArgs struct {
	cnitypes.CommonArgs
	IP net.IP `json:"ip,omitempty"`
}

func cmdAdd(args *skel.CmdArgs) error {
	// types.NetConf 也就是cni配置文件里的内容，具体内容可见第一篇文章
	conf := types.NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	// 准备好相关参数
	nodename := utils.DetermineNodename(conf)
	utils.ConfigureLogging(conf)
	calicoClient, err := utils.CreateClient(conf)
	epIDs, err := utils.GetIdentifiers(args, nodename)
	epIDs.WEPName, err = epIDs.CalculateWorkloadEndpointName(false)
	handleID := utils.GetHandleID(conf.Name, args.ContainerID, epIDs.WEPName)
	ipamArgs := ipamArgs{}
	if err = cnitypes.LoadArgs(args.Args, &ipamArgs); err != nil {
		return err
	}

	r := &current.Result{}
	if ipamArgs.IP != nil {
        // 这里分配指定IP，我们创建pod并没有通过annotation指定IP，而且一般都没有去指定
		// ...
	} else {
		// 没有指定IP，让calico-ipam帮我们从节点的pod cidr里去分配一个IP，我们生产calico会走这个逻辑

        // 这里如果cni配置文件没有指定conf.IPAM.IPv4Pools，则从calico datastore数据库查询可以使用的ippool
        // ippool是calico在启动时就已经写入数据库的，值是可以我们根据生产环境配置的
        // 因为会从这个ippool，即集群大网段cluster cidr切分出节点子网段node cidr，再从node cidr中allocate出一个pod ip地址，
        // 所以先查询出我们集群的ippool是什么
		v4pools, err := utils.ResolvePools(ctx, calicoClient, conf.IPAM.IPv4Pools, true)
		var maxBlocks int
		assignArgs := ipam.AutoAssignArgs{
			Num4:             num4,
			Num6:             num6,
			HandleID:         &handleID,
			Hostname:         nodename,
			IPv4Pools:        v4pools,
			IPv6Pools:        v6pools,
			MaxBlocksPerHost: maxBlocks,
			Attrs:            attrs,
		}
		
		autoAssignWithLock := func(calicoClient client.Interface, ctx context.Context, assignArgs ipam.AutoAssignArgs) ([]cnet.IPNet, []cnet.IPNet, error) {
			// ...
			// 这里会调用IPAM模块，来从node cidr中随机分配一个还未分配的IP地址
			return calicoClient.IPAM().AutoAssign(ctx, assignArgs)
		}
		assignedV4, assignedV6, err := autoAssignWithLock(calicoClient, ctx, assignArgs)
	}

	// Print result to stdout, in the format defined by the requested cniVersion.
	return cnitypes.PrintResult(r, conf.CNIVersion)
}

```

以上代码重点是调用IPAM模块的AutoAssign()函数来自动分配IP地址，看下 **[AutoAssign()](https://github.com/projectcalico/libcalico-go/blob/release-v3.17/lib/ipam/ipam.go#L80-L127)** 代码，
代码在 **projectcalico/libcalico-go** 代码仓库里，该仓库作为公共基础仓库，被 **projectcalico/cni-plugin** 和 **projectcalico/calicoctl** 等仓库引用：

```go

// 从AutoAssignArgs.IPv4Pools中自动分配一个IP
func (c ipamClient) AutoAssign(ctx context.Context, args AutoAssignArgs) ([]net.IPNet, []net.IPNet, error) {
	hostname, err := decideHostname(args.Hostname)
	// ...
	if args.Num4 != 0 {
		v4list, err = c.autoAssign(ctx, args.Num4, args.HandleID, args.Attrs, args.IPv4Pools, 4, hostname, args.MaxBlocksPerHost, args.HostReservedAttrIPv4s)
	}
	// ...
	return v4list, v6list, nil
}

```










## 总结







## 参考文献







