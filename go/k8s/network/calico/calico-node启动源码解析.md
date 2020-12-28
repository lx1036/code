

# Kubernetes学习笔记之Calico Startup源码解析

## Overview
我们目前生产k8s和calico使用ansible二进制部署在私有机房，没有使用官方的calico/node容器部署，并且因为没有使用network policy只部署了confd/bird进程服务，没有部署felix。
采用BGP(Border Gateway Protocol)方式来部署网络，并且采用 **[Peered with TOR (Top of Rack) routers](https://docs.projectcalico.org/networking/determine-best-networking#on-prem)** 
方式部署，每一个worker node和其置顶交换机建立bgp peer配对，置顶交换机会继续和上层核心交换机建立bgp peer配对，这样可以保证pod ip在公司内网可以直接被访问。

> BGP: 主要是网络之间分发动态路由的一个协议，使用TCP协议传输数据。比如，交换机A下连着12台worker node，可以在每一台worker node上安装一个BGP Client，如Bird或GoBGP程序，
> 这样每一台worker node会把自己的路由分发给交换机A，交换机A会做路由聚合，以及继续向上一层核心交换机转发。交换机A上的路由是Node级别，而不是Pod级别的。

平时在维护k8s云平台时，有时发现一台worker节点上的所有pod ip在集群外没法访问，经过排查发现是该worker节点有两张内网网卡eth0和eth1，eth0 IP地址和交换机建立BGP
连接，并获取其as number号，但是bird启动配置文件bird.cfg里使用的eth1网卡IP地址。并且发现calico里的 **[Node](https://docs.projectcalico.org/reference/resources/node)** 
数据的IP地址ipv4Address和 **[BGPPeer](https://docs.projectcalico.org/reference/resources/bgppeer)** 数据的交换机地址peerIP也对不上。可以通过如下命令获取calico数据：

```shell

calicoctl get node ${nodeName} -o yaml
calicoctl get bgppeer ${peerName} -o yaml

```

一番抓头挠腮后，找到根本原因是我们的ansible部署时，在调用网络API获取交换机的bgp peer的as number和peer ip数据时，使用的是eth0地址，
并且通过ansible任务`calicoctl apply -f bgp_peer.yaml` 写入 **[Node-specific BGP Peer](https://docs.projectcalico.org/reference/resources/bgppeer#node-specific-peer)**数据，
写入calico BGP Peer数据里使用的是eth0交换机地址。但是ansible任务跑到配置bird.cfg配置文件时，环境变量IP使用的是eth1 interface，
写入calico Node数据使用的是eth1网卡地址，然后被confd进程读取Node数据生成bird.cfg文件时，使用的就会是eth1网卡地址。这里应该是使用eth0才对。

找到问题原因后，就愉快的解决了。

但是，又突然想知道，calico是怎么写入Node数据的？代码原来在calico启动代码 **[startup.go](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go)** 这里。
官方提供的calico/node容器里，会启动bird/confd/felix等多个进程，并且使用runsvdir(类似supervisor)来管理多个进程。容器启动时，也会进行运行初始化脚本，
配置在这里 **[L11-L13](https://github.com/projectcalico/node/blob/release-v3.17/filesystem/etc/rc.local#L11-L13)** :

```shell

# Run the startup initialisation script.
# These ensure the node is correctly configured to run.
calico-node -startup || exit 1

```

所以，可以看下初始化脚本做了什么工作。

## 初始化脚本源码解析
当运行`calico-node -startup`命令时，实际上会执行 **[L111-L113](https://github.com/projectcalico/node/blob/release-v3.17/cmd/calico-node/main.go#L111-L113)** ，
也就是starup模块下的startup.go脚本:

```go

  func main() {
    // ...
    if *runStartup {
        logrus.SetFormatter(&logutils.Formatter{Component: "startup"})
        startup.Run()
    }
    // ...
  }
  
```

startup.go脚本主要做了三件事情 **[L91-L96](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go#L91-L96)** ：
* Detecting IP address and Network to use for BGP.
* Configuring the node resource with IP/AS information provided in the environment, or autodetected.
* Creating default IP Pools for quick-start use.(可以通过NO_DEFAULT_POOLS关闭，一个集群就只需要一个IP Pool，
  不需要每一次初始化都去创建一次。不过官方代码里已经适配了如果集群内有IP Pool，可以跳过创建，所以也可以不关闭。我们生产k8s ansible部署这里是选择关闭，不关闭也不影响)

所以，初始化时只做一件事情：往calico里写入一个Node数据，供后续confd配置bird.cfg配置使用。看一下启动脚本具体执行逻辑 **[L97-L223](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go#L97-L223)** ：

```go

func Run() {
  // ...
  // 从NODENAME、HOSTNAME等环境变量或者CALICO_NODENAME_FILE文件内，读取当前宿主机名字
  nodeName := determineNodeName()
  
  // 创建CalicoClient: 
  // 如果DATASTORE_TYPE使用kubernetes，只需要传KUBECONFIG变量值就行，如果k8s pod部署，都不需要传，这样就和创建
  // KubernetesClient一样道理，可以参考calicoctl的配置文档：https://docs.projectcalico.org/getting-started/clis/calicoctl/configure/kdd
  // 如果DATASTORE_TYPE使用etcdv3，还得配置etcd相关的环境变量值，可以参考: https://docs.projectcalico.org/getting-started/clis/calicoctl/configure/etcd
  // 平时本地编写calico测试代码时，可以在~/.zshrc里加上环境变量，可以参考 https://docs.projectcalico.org/getting-started/clis/calicoctl/configure/kdd#example-using-environment-variables :
  // export CALICO_DATASTORE_TYPE=kubernetes
  // export CALICO_KUBECONFIG=~/.kube/config
  cfg, cli := calicoclient.CreateClient()
  // ...
  if os.Getenv("WAIT_FOR_DATASTORE") == "true" {
    // 通过c.Nodes.Get("foo")来测试下是否能正常调用
    waitForConnection(ctx, cli)
  }
  // ...

  // 从calico中查询nodeName的Node数据，如果没有则构造个新Node对象
  // 后面会用该宿主机的IP地址来更新该Node对象
  node := getNode(ctx, cli, nodeName)

  var clientset *kubernetes.Clientset
  var kubeadmConfig, rancherState *v1.ConfigMap

  // If running under kubernetes with secrets to call k8s API
  if config, err := rest.InClusterConfig(); err == nil {
    // 如果是kubeadm或rancher部署的k8s集群，读取kubeadm-config或full-cluster-state ConfigMap值
    // 为后面配置ClusterType变量以及创建IPPool使用
    // 我们生产k8s目前没使用这两种方式
    
    // ...
  }

  // 这里逻辑是关键，这里会配置Node对象的spec.bgp.ipv4Address地址，而且获取ipv4地址策略多种方式
  // 可以直接给IP环境变量自己指定一个具体地址如10.203.10.20，也可以给IP环境变量指定"autodetect"自动检测
  // 而自动检测策略是根据"IP_AUTODETECTION_METHOD"环境变量配置的，有can-reach或interface=eth.*等等，
  // 具体自动检测策略可以参考：https://docs.projectcalico.org/archive/v3.17/networking/ip-autodetection
  // 我们的生产k8s是在ansible里根据变量获取eth{$interface}的ipv4地址给IP环境变量，而如果机器是双内网网卡，不管是选择eth0还是eth1地址
  // 要和创建bgp peer时使用的网卡要保持一致，另外还得看这台机器默认网关地址是eth0还是eth1的默认网关
  // 有关具体如何获取IP地址，下文详解
  configureAndCheckIPAddressSubnets(ctx, cli, node)

  // 我们使用bird，这里CALICO_NETWORKING_BACKEND配置bird
  if os.Getenv("CALICO_NETWORKING_BACKEND") != "none" {
    // 这里从环境变量AS中查询，可以给个默认值65188，不影响
    configureASNumber(node)
    if clientset != nil {
      // 如果是选择官方那种calico/node集群内部署，这里会patch下k8s的当前Node的 NetworkUnavailable Condition，意思是网络当前不可用
      // 可以参考https://kubernetes.io/docs/concepts/architecture/nodes/#condition
      // 目前我们生产k8s没有calico/node集群内部署，所以不会走这一步逻辑，并且我们生产k8s版本过低，Node Conditions里也没有NetworkUnavailable Condition
      err := setNodeNetworkUnavailableFalse(*clientset, nodeName)
      // ...
    }
  }
  
  // 配置下node.Spec.OrchRefs为k8s，值从CALICO_K8S_NODE_REF环境变量里读取
  configureNodeRef(node)
  // 创建/var/run/calico、/var/lib/calico和/var/log/calico等目录
  ensureFilesystemAsExpected()
  
  // calico Node对象已经准备好了，可以创建或更新Node对象
  // 这里是启动脚本的最核心逻辑，以上都是为了查询Node对象相关的配置数据，主要作用就是为了初始化时创建或更新Node对象
  if _, err := CreateOrUpdate(ctx, cli, node); err != nil {
    // ...
  }

  // 配置集群的IP Pool，即整个集群的pod cidr网段，如果使用/18网段，每一个k8s worker Node使用/27子网段，那就是集群最多可以部署2^(27-18)=512
  // 台机器，每台机器可以分配2^(32-27)=32-首位两个地址=30个pod。
  configureIPPools(ctx, cli, kubeadmConfig)

  // 这里主要写一个名字为default的全局FelixConfiguration对象，以及DatastoreType不是kubernetes，就会对于每一个Node写一个该Node的
  // 默认配置的FelixConfiguration对象。
  // 我们生产k8s使用etcdv3，所以初始化时会看到calico数据里会有每一个Node的FelixConfiguration对象。另外，我们没使用felix，不需要太关注felix数据。
  if err := ensureDefaultConfig(ctx, cfg, cli, node, getOSType(), kubeadmConfig, rancherState); err != nil {
    log.WithError(err).Errorf("Unable to set global default configuration")
    terminate()
  }

  // 把nodeName写到CALICO_NODENAME_FILE环境变量指定的文件内
  writeNodeConfig(nodeName)
  // ...
}
// 从calico中查询nodeName的Node数据，如果没有则构造个新Node对象
func getNode(ctx context.Context, client client.Interface, nodeName string) *api.Node {
    node, err := client.Nodes().Get(ctx, nodeName, options.GetOptions{})
    // ...
    if err != nil {
      // ...
        node = api.NewNode()
        node.Name = nodeName
    }
    return node
}
// 创建或更新Node对象
func CreateOrUpdate(ctx context.Context, client client.Interface, node *api.Node) (*api.Node, error) {
    if node.ResourceVersion != "" {
        return client.Nodes().Update(ctx, node, options.SetOptions{})
    }
    return client.Nodes().Create(ctx, node, options.SetOptions{})
}

```

通过上面代码分析，有两个关键逻辑需要仔细看下：一个是获取当前机器的IP地址；一个是配置集群的pod cidr。

这里先看下配置集群pod cidr逻辑 **[L858-L1050](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go#L858-L1050)** ：

```go

// configureIPPools ensures that default IP pools are created (unless explicitly requested otherwise).
func configureIPPools(ctx context.Context, client client.Interface, kubeadmConfig *v1.ConfigMap) {
  // Read in environment variables for use here and later.
  ipv4Pool := os.Getenv("CALICO_IPV4POOL_CIDR")
  ipv6Pool := os.Getenv("CALICO_IPV6POOL_CIDR")

  if strings.ToLower(os.Getenv("NO_DEFAULT_POOLS")) == "true" {
    // ...
    return
  }
  // ...
  // 从CALICO_IPV4POOL_BLOCK_SIZE环境变量中读取block size，即你的网段要分配的子网段掩码是多少，比如这里默认值是/26
  // 如果选择默认的192.168.0.0/16 ip pool，而分配给每个Node子网是/26网段，那集群可以部署2^(26-16)=1024台机器了
  ipv4BlockSizeEnvVar := os.Getenv("CALICO_IPV4POOL_BLOCK_SIZE")
  if ipv4BlockSizeEnvVar != "" {
    ipv4BlockSize = parseBlockSizeEnvironment(ipv4BlockSizeEnvVar)
  } else {
    // DEFAULT_IPV4_POOL_BLOCK_SIZE为默认26子网段
    ipv4BlockSize = DEFAULT_IPV4_POOL_BLOCK_SIZE
  }
  // ...
  // Get a list of all IP Pools
  poolList, err := client.IPPools().List(ctx, options.ListOptions{})
  // ...
  // Check for IPv4 and IPv6 pools.
  ipv4Present := false
  ipv6Present := false
  for _, p := range poolList.Items {
    ip, _, err := cnet.ParseCIDR(p.Spec.CIDR)
    if err != nil {
      log.Warnf("Error parsing CIDR '%s'. Skipping the IPPool.", p.Spec.CIDR)
    }
    version := ip.Version()
    ipv4Present = ipv4Present || (version == 4)
    ipv6Present = ipv6Present || (version == 6)
    // 这里官方做了适配，如果集群内有ip pool，后面逻辑就不会调用createIPPool()创建ip pool
    if ipv4Present && ipv6Present {
      break
    }
  }
  if ipv4Pool == "" {
    // 如果没配置pod网段，给个默认网段"192.168.0.0/16"
    ipv4Pool = DEFAULT_IPV4_POOL_CIDR
        // ...
  }
  // ...
  // 集群内已经有ip pool，这里就不会重复创建
  if !ipv4Present {
    log.Debug("Create default IPv4 IP pool")
    outgoingNATEnabled := evaluateENVBool("CALICO_IPV4POOL_NAT_OUTGOING", true)

    createIPPool(ctx, client, ipv4Cidr, DEFAULT_IPV4_POOL_NAME, ipv4IpipModeEnvVar, ipv4VXLANModeEnvVar, outgoingNATEnabled, ipv4BlockSize, ipv4NodeSelector)
  }
  // ... 省略ipv6逻辑
}

// 创建ip pool
func createIPPool(ctx context.Context, client client.Interface, cidr *cnet.IPNet, poolName, ipipModeName, vxlanModeName string, isNATOutgoingEnabled bool, blockSize int, nodeSelector string) {
  //...
  pool := &api.IPPool{
    ObjectMeta: metav1.ObjectMeta{
      Name: poolName,
    },
    Spec: api.IPPoolSpec{
      CIDR:         cidr.String(),
      NATOutgoing:  isNATOutgoingEnabled,
      IPIPMode:     ipipMode, // 因为我们生产使用bgp，这里ipipMode值是never
      VXLANMode:    vxlanMode,
      BlockSize:    blockSize,
      NodeSelector: nodeSelector,
    },
  }
  // 创建ip pool
  if _, err := client.IPPools().Create(ctx, pool, options.SetOptions{}); err != nil {
    // ...
  }
}

```

然后看下自动获取IP地址的逻辑 **[L498-L585](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go#L498-L585)** ：

```go

// 给Node对象配置IPv4Address地址
func configureIPsAndSubnets(node *api.Node) (bool, error) {
  // ...
  oldIpv4 := node.Spec.BGP.IPv4Address

  // 从IP环境变量获取IP地址，我们生产k8s ansible直接读取的网卡地址，但是对于双内网网卡，有时这里读取IP地址时，
  // 会和bgp_peer.yaml里采用的IP地址会不一样，我们目前生产的bgp_peer.yaml里默认采用eth0的地址，写死的(因为我们机器网关地址默认都是eth0的网关)，
  // 所以这里的IP一定得是eth0的地址。
  ipv4Env := os.Getenv("IP")
  if ipv4Env == "autodetect" || (ipv4Env == "" && node.Spec.BGP.IPv4Address == "") {
    adm := os.Getenv("IP_AUTODETECTION_METHOD")
    // 这里根据自动检测策略来判断选择哪个网卡地址，比较简单不赘述，可以看代码 **[L701-L746](https://github.com/projectcalico/node/blob/release-v3.17/pkg/startup/startup.go#L701-L746)** 
    // 和配置文档 **[ip-autodetection](https://docs.projectcalico.org/archive/v3.17/networking/ip-autodetection)** ，
    // 如果使用calico/node在k8s内部署，根据一些讨论言论，貌似使用can-reach=xxx可以少踩很多坑
    cidr := autoDetectCIDR(adm, 4)
    if cidr != nil {
      // We autodetected an IPv4 address so update the value in the node.
      node.Spec.BGP.IPv4Address = cidr.String()
    } else if node.Spec.BGP.IPv4Address == "" {
      return false, fmt.Errorf("Failed to autodetect an IPv4 address")
    } else {
      // ...
    }
  } else if ipv4Env == "none" && node.Spec.BGP.IPv4Address != "" {
    log.Infof("Autodetection for IPv4 disabled, keeping existing value: %s", node.Spec.BGP.IPv4Address)
    validateIP(node.Spec.BGP.IPv4Address)
  } else if ipv4Env != "none" {
    // 我们生产k8s ansible走的是这个逻辑，而且直接取的是eth0的IP地址，subnet会默认被设置为/32
    // 可以参考官网文档：https://docs.projectcalico.org/archive/v3.17/networking/ip-autodetection#manually-configure-ip-address-and-subnet-for-a-node
    if ipv4Env != "" {
      node.Spec.BGP.IPv4Address = parseIPEnvironment("IP", ipv4Env, 4)
    }
    validateIP(node.Spec.BGP.IPv4Address)
  }
  // ...
  // Detect if we've seen the IP address change, and flag that we need to check for conflicting Nodes
  if node.Spec.BGP.IPv4Address != oldIpv4 {
    log.Info("Node IPv4 changed, will check for conflicts")
    return true, nil
  }

  return false, nil
}

```

以上就是calico启动脚本执行逻辑，比较简单，但是学习了其代码逻辑之后，对问题排查会更加得心应手，否则只能傻瓜式的乱猜，
尽管碰巧解决了问题但是不知道为什么，后面再次遇到类似问题还是不知道怎么解决，浪费时间。


## 总结
本文主要学习了下calico启动脚本执行逻辑，主要是往calico里写部署宿主机的Node数据，容易出错的地方是机器双网卡时可能会出现Node和BGPPeer数据不一致，
bird没法分发路由，导致该机器的pod地址没法集群外和集群内被路由到。

目前我们生产calico用的ansible二进制部署，通过日志排查也不方便，还是推荐calico/node容器化部署在k8s内，调用网络API与交换机bgp peer配对时，获取相关数据逻辑，
可以放在initContainers里，然后`calicoctl apply -f bgp_peer.yaml`写到calico里。当然，不排除中间会踩不少坑，以及时间精力问题。

总之，calico是一个优秀的k8s cni实现，使用成熟方案BGP协议来分发路由，数据包走三层路由且中间没有SNAT/DNAT操作，也非常容易理解其原理过程。
后续，会写一写kubelet在创建sandbox容器的network namespace时，如何调用calico命令来创建相关网络对象和网卡，以及使用calico-ipam来分配当前Node节点的子网段和给pod
分配ip地址。
