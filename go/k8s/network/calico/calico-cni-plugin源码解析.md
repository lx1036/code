# Calico疑惑
## Overview
IP地址分配的性能有哪些问题要考虑呢？在大规模集群的场景下，Calico IP地址的分配速率是否受到集群规模的限制?
IP地址和Block size怎么配置才能保持高速的IP地址分配?另外，Calico的IP地址在Node节点异常时，IP地址如何回收?
什么时候有可能产生IP地址冲突？为了解答这些疑问，需要熟悉CalicoIP地址分配的执行流程。

debug调试calico cni插件，或者查看cni日志，记得打开debug level日志：
```shell
sudo journalctl --since="2020-12-27 01:04:00" -r -u kubelet
```

## 参考文献
**[Use a specific IP address with a pod](https://docs.projectcalico.org/networking/use-specific-ip)**
**[Calico IPAM源码解析](https://mp.weixin.qq.com/s/lyfeZh6VWWjXuLY8fl3ciw)**
**[calico,CNI的一种实现](https://www.yuque.com/baxiaoshi/tyado3/lvfa0b)**
**[containernetworking/cni](https://github.com/containernetworking/cni)**
**[projectcalico/cni-plugin](https://github.com/projectcalico/cni-plugin)**



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







