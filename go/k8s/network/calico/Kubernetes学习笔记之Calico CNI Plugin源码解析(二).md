

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
**[Use a specific IP address with a pod](https://docs.projectcalico.org/networking/use-specific-ip)**
**[Calico IPAM源码解析](https://mp.weixin.qq.com/s/lyfeZh6VWWjXuLY8fl3ciw)**
**[calico,CNI的一种实现](https://www.yuque.com/baxiaoshi/tyado3/lvfa0b)**
**[containernetworking/cni](https://github.com/containernetworking/cni)**
**[projectcalico/cni-plugin](https://github.com/projectcalico/cni-plugin)**

