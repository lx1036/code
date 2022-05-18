package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexflint/go-filemutex"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/cmd/cni-plugin/driver"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/link"
	"k8s.io/apimachinery/pkg/util/wait"
	"net"
	"path/filepath"
	"runtime"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/rpc"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	"github.com/containernetworking/cni/pkg/skel"
	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/cni/pkg/version"
	cniversion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

const (
	defaultMTU = 1500

	defaultCniTimeout = 120 * time.Second
	defaultSocketPath = "/var/run/eni/eni.socket"
	defaultCNILock    = "/var/run/eni/eni.lock"

	defaultEventTimeout = 10 * time.Second

	defaultVethForENI = "veth1"
	defaultVethPrefix = "lxc"

	TypeENIMultiIP_IPVLAN      = "ipvlan"
	TypeENIMultiIP_POLICYROUTE = "policyRoute"
)

var (
	veth, rawNIC, ipvlan driver.NetnsDriver

	LinkIP   = net.IPv4(169, 254, 1, 1)
	LinkIPv6 = net.ParseIP("fe80::1")
)

// NetConf is the cni network config
type NetConf struct {
	cniTypes.NetConf

	// HostVethPrefix is the veth for container prefix on host
	HostVethPrefix string `json:"veth_prefix"`

	// eniIPVirtualType is the ipvlan for container
	ENIIPVirtualType string `json:"eniip_virtual_type"`

	// HostStackCIDRs is a list of CIDRs, all traffic targeting these CIDRs will be redirected to host network stack
	HostStackCIDRs []string `json:"host_stack_cidrs"`

	// MTU is container and ENI network interface MTU
	MTU int `json:"mtu"`

	EnableDebug bool `json:"enable-debug"`
}

// K8SArgs is cni args of kubernetes, @see https://github.com/kubernetes/kubernetes/blob/v1.19.7/pkg/kubelet/dockershim/network/cni/cni.go#L400-L405
type K8SArgs struct {
	cniTypes.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               cniTypes.UnmarshallableString // nolint
	K8S_POD_NAMESPACE          cniTypes.UnmarshallableString // nolint
	K8S_POD_INFRA_CONTAINER_ID cniTypes.UnmarshallableString // nolint
}

func init() {
	// This ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, "eni")
}

func cmdAdd(args *skel.CmdArgs) error {
	confVersion, cniNetns, conf, k8sConfig, err := parseCmdArgs(args)
	if err != nil {
		return err
	}
	defer cniNetns.Close()

	klog.Infof(fmt.Sprintf("args: %+v", *args))

	eniBackendClient, closeConn, err := getNetworkClient()
	if err != nil {
		return fmt.Errorf("error create grpc client,pod %s/%s, %w", string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME), err)
	}
	defer closeConn()

	timeoutContext, cancel := context.WithTimeout(context.Background(), defaultCniTimeout)
	defer cancel()

	defer func() {
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), defaultEventTimeout)
			defer cancel()
			_, _ = eniBackendClient.RecordEvent(ctx,
				&rpc.EventRequest{
					EventTarget:     rpc.EventTarget_EventTargetPod,
					K8SPodName:      string(k8sConfig.K8S_POD_NAME),
					K8SPodNamespace: string(k8sConfig.K8S_POD_NAMESPACE),
					EventType:       rpc.EventType_EventTypeWarning,
					Reason:          "AllocateIPFailed",
					Message:         err.Error(),
				})
		}
	}()

	// INFO: (1) allocate IP
	allocateIPReply, err := eniBackendClient.AllocateIP(
		timeoutContext,
		&rpc.AllocateIPRequest{
			Netns:                  args.Netns,
			K8SPodName:             string(k8sConfig.K8S_POD_NAME),
			K8SPodNamespace:        string(k8sConfig.K8S_POD_NAMESPACE),
			K8SPodInfraContainerId: string(k8sConfig.K8S_POD_INFRA_CONTAINER_ID),
			IfName:                 args.IfName,
		})
	if err != nil {
		return fmt.Errorf("cmdAdd: error alloc ip for pod %s: %v", fmt.Sprintf("%s/%s", string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME)), err)
	}
	if !allocateIPReply.Success {
		return fmt.Errorf("cmdAdd: error alloc ip for pod %s", fmt.Sprintf("%s/%s", string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME)))
	}
	klog.Infof(fmt.Sprintf("%#v", allocateIPReply))
	defer func() {
		if err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), defaultCniTimeout)
			defer cancel()
			_, _ = eniBackendClient.ReleaseIP(ctx,
				&rpc.ReleaseIPRequest{
					K8SPodName:             string(k8sConfig.K8S_POD_NAME),
					K8SPodNamespace:        string(k8sConfig.K8S_POD_NAMESPACE),
					K8SPodInfraContainerId: string(k8sConfig.K8S_POD_INFRA_CONTAINER_ID),
					IPType:                 allocateIPReply.IPType,
					Reason:                 fmt.Sprintf("roll back ip for error: %v", err),
				})
		}
	}()

	hostIPSet, err := driver.GetHostIP(true, false) // 宿主机 eth0 地址
	if err != nil {
		return err
	}
	hostVethIfName, err := link.VethNameForPod(string(k8sConfig.K8S_POD_NAME), string(k8sConfig.K8S_POD_NAMESPACE), defaultVethPrefix)
	if err != nil {
		return err
	}

	var containerIPNet *types.IPNetSet
	var gatewayIPSet *types.IPSet
	for _, netConf := range allocateIPReply.NetConfs {
		var setupCfg *types.SetupConfig
		setupCfg, err = parseSetupConf(args, netConf, conf, allocateIPReply.IPType)
		setupCfg.HostVethIfName = hostVethIfName
		setupCfg.HostIPSet = hostIPSet

		switch allocateIPReply.IPType {

		case rpc.IPType_TypeENIMultiIP: // ipvlan or policyRoute
			eniMultiIPType := TypeENIMultiIP_IPVLAN
			if setupCfg.ContainerIfName == args.IfName {
				containerIPNet = setupCfg.ContainerIPNet
				gatewayIPSet = setupCfg.GatewayIP
			}

			switch eniMultiIPType {
			case TypeENIMultiIP_IPVLAN:
				err = driver.NewIPVlanDriver().Setup(setupCfg, cniNetns) // ipvlan 模式

			case TypeENIMultiIP_POLICYROUTE:
				err = driver.NewPolicyRoute().Setup(setupCfg, cniNetns) // veth 策略路由模式
			}

			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not support this network type")
		}
	}

	index := 0
	result := &current.Result{}
	result.Interfaces = append(result.Interfaces, &current.Interface{
		Name: args.IfName,
	})
	if containerIPNet.IPv4 != nil && gatewayIPSet.IPv4 != nil {
		result.IPs = append(result.IPs, &current.IPConfig{
			Address:   *containerIPNet.IPv4,
			Gateway:   gatewayIPSet.IPv4,
			Interface: &index,
		})
	}
	if containerIPNet.IPv6 != nil && gatewayIPSet.IPv6 != nil {
		result.IPs = append(result.IPs, &current.IPConfig{
			Address:   *containerIPNet.IPv6,
			Gateway:   gatewayIPSet.IPv6,
			Interface: &index,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultEventTimeout)
	defer cancel()
	_, _ = eniBackendClient.RecordEvent(ctx,
		&rpc.EventRequest{
			EventTarget:     rpc.EventTarget_EventTargetPod,
			K8SPodName:      string(k8sConfig.K8S_POD_NAME),
			K8SPodNamespace: string(k8sConfig.K8S_POD_NAMESPACE),
			EventType:       rpc.EventType_EventTypeNormal,
			Reason:          "AllocateIPSucceed",
			Message:         fmt.Sprintf("Allocate IP %s for Pod", containerIPNet.String()),
		})

	return cniTypes.PrintResult(result, confVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	_, cniNetns, conf, k8sConfig, err := parseCmdArgs(args)
	if err != nil {
		return err
	}
	defer cniNetns.Close()

	klog.Infof(fmt.Sprintf("args: %+v", *args))

	eniBackendClient, closeConn, err := getNetworkClient()
	if err != nil {
		return fmt.Errorf("error create grpc client,pod %s/%s, %w", string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME), err)
	}
	defer closeConn()

	timeoutContext, cancel := context.WithTimeout(context.Background(), defaultCniTimeout)
	defer cancel()

	ipInfoReply, err := eniBackendClient.GetIPInfo(timeoutContext, &rpc.GetInfoRequest{
		K8SPodName:             string(k8sConfig.K8S_POD_NAME),
		K8SPodNamespace:        string(k8sConfig.K8S_POD_NAMESPACE),
		K8SPodInfraContainerId: string(k8sConfig.K8S_POD_INFRA_CONTAINER_ID),
	})
	if err != nil {
		return fmt.Errorf("error get ip from terway, pod %s/%s, %w", string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME), err)
	}

	// 文件锁
	lock, err := GetFileLock(defaultCNILock)
	if err != nil {
		return err
	}
	defer lock.Close()

	for _, netConf := range ipInfoReply.NetConfs {
		teardownCfg, err := parseTeardownConf(args, netConf, conf, ipInfoReply.IPType)
		if err != nil {
			return err
		}
		switch ipInfoReply.IPType {
		case rpc.IPType_TypeENIMultiIP: // ipvlan or policyRoute
			err = driver.NewIPVlanDriver().Teardown(teardownCfg, cniNetns) // ipvlan 模式
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("not support this network type")
		}
	}

	releaseIPReply, err := eniBackendClient.ReleaseIP(timeoutContext, &rpc.ReleaseIPRequest{
		K8SPodName:             string(k8sConfig.K8S_POD_NAME),
		K8SPodNamespace:        string(k8sConfig.K8S_POD_NAMESPACE),
		K8SPodInfraContainerId: string(k8sConfig.K8S_POD_INFRA_CONTAINER_ID),
		Reason:                 "normal release",
	})
	if err != nil || !releaseIPReply.GetSuccess() {
		return fmt.Errorf("error release ip for pod, maybe cause resource leak: %v, %s", err, releaseIPReply.String())
	}
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	panic("not implemented")
}

func parseCmdArgs(args *skel.CmdArgs) (string, ns.NetNS, *NetConf, *K8SArgs, error) {
	versionDecoder := &cniversion.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(args.StdinData)
	if err != nil {
		return "", nil, nil, nil, err
	}
	netNS, err := ns.GetNS(args.Netns) // open fd /proc/{pid}/ns/net
	if err != nil {
		return "", nil, nil, nil, err
	}

	conf := NetConf{}
	if err = json.Unmarshal(args.StdinData, &conf); err != nil {
		return "", nil, nil, nil, fmt.Errorf("error parse args, %w", err)
	}
	if conf.MTU == 0 {
		conf.MTU = defaultMTU
	}

	k8sConfig := K8SArgs{}
	if err = cniTypes.LoadArgs(args.Args, &k8sConfig); err != nil {
		return "", nil, nil, nil, fmt.Errorf("error parse args, %w", err)
	}

	return confVersion, netNS, &conf, &k8sConfig, nil
}

// INFO: 获取 ENI gRPC Client，调用 ENI
func getNetworkClient() (rpc.EniBackendClient, func(), error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), defaultCniTimeout)
	grpcConn, err := grpc.DialContext(timeoutCtx, defaultSocketPath, grpc.WithInsecure(), grpc.WithContextDialer(
		func(ctx context.Context, s string) (net.Conn, error) {
			unixAddr, err := net.ResolveUnixAddr("unix", defaultSocketPath)
			if err != nil {
				return nil, fmt.Errorf("error while resolve unix addr:%w", err)
			}
			d := &net.Dialer{}
			return d.DialContext(timeoutCtx, "unix", unixAddr.String())
		}))
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("error dial to terway %s, %w", defaultSocketPath, err)
	}

	eniBackendClient := rpc.NewEniBackendClient(grpcConn)
	return eniBackendClient, func() {
		grpcConn.Close()
		cancel()
	}, nil
}

func parseSetupConf(args *skel.CmdArgs, alloc *rpc.NetConf, conf *types.CNIConf, ipType rpc.IPType) (*types.SetupConfig, error) {
	var (
		err            error
		containerIPNet *types.IPNetSet
		gatewayIPSet   *types.IPSet
		serviceCIDR    *types.IPNetSet
		eniGatewayIP   *types.IPSet
		deviceID       int32
		trunkENI       bool
		vid            uint32

		ingress uint64
		egress  uint64

		routes []cniTypes.Route

		disableCreatePeer bool
	)

	serviceCIDR, err = types.ToIPNetSet(alloc.GetBasicInfo().GetServiceCIDR())
	if err != nil {
		return nil, err
	}

	podIP := alloc.GetBasicInfo().GetPodIP()
	subNet := alloc.GetBasicInfo().GetPodCIDR()
	gatewayIP := alloc.GetBasicInfo().GetGatewayIP()
	containerIPNet, err = types.BuildIPNet(podIP, subNet)
	if err != nil {
		return nil, err
	}
	gatewayIPSet, err = types.ToIPSet(gatewayIP)
	if err != nil {
		return nil, err
	}
	disableCreatePeer = conf.DisableHostPeer

	if alloc.GetENIInfo() != nil {
		mac := alloc.GetENIInfo().GetMAC()
		if mac != "" {
			deviceID, err = link.GetDeviceNumberByMac(mac)
			if err != nil {
				return nil, err
			}
		}
		trunkENI = alloc.GetENIInfo().GetTrunk()
		vid = alloc.GetENIInfo().GetVid()
		if alloc.GetENIInfo().GetGatewayIP() != nil {
			eniGatewayIP, err = types.ToIPSet(alloc.GetENIInfo().GetGatewayIP())
			if err != nil {
				return nil, err
			}
		}
	}

	if alloc.GetPod() != nil {
		ingress = alloc.GetPod().GetIngress()
		egress = alloc.GetPod().GetEgress()
	}

	hostStackCIDRs := make([]*net.IPNet, 0)
	for _, v := range conf.HostStackCIDRs {
		_, cidr, err := net.ParseCIDR(v)
		if err != nil {
			return nil, fmt.Errorf("host_stack_cidrs(%s) is invaild: %v", v, err)
		}
		hostStackCIDRs = append(hostStackCIDRs, cidr)
	}

	containerIfName := alloc.IfName
	if containerIfName == "" {
		containerIfName = args.IfName
	}

	for _, r := range alloc.GetExtraRoutes() {
		ip, n, err := net.ParseCIDR(r.Dst)
		if err != nil {
			return nil, fmt.Errorf("error parse extra routes, %w", err)
		}
		route := cniTypes.Route{Dst: *n}
		if ip.To4() != nil {
			route.GW = gatewayIPSet.IPv4
		} else {
			route.GW = gatewayIPSet.IPv6
		}
		routes = append(routes, route)
	}

	return &types.SetupConfig{
		ContainerIfName: containerIfName,
		ContainerIPNet:  containerIPNet,
		GatewayIP:       gatewayIPSet,

		MTU:          conf.MTU,
		ENIIndex:     int(deviceID), // 弹性网卡 interface index in host network namespace
		ENIGatewayIP: eniGatewayIP,

		DisableCreatePeer: disableCreatePeer,
		StripVlan:         trunkENI,
		Vid:               int(vid), // vlan id
		DefaultRoute:      alloc.GetDefaultRoute(),

		MultiNetwork: false,
		ExtraRoutes:  routes,

		ServiceCIDR:    serviceCIDR,
		HostStackCIDRs: hostStackCIDRs,

		Ingress: ingress,
		Egress:  egress,
	}, nil
}

func parseTeardownConf(args *skel.CmdArgs, alloc *rpc.NetConf, conf *types.CNIConf, ipType rpc.IPType) (*types.TeardownCfg, error) {
	if alloc.GetBasicInfo() == nil {
		return nil, fmt.Errorf("return empty pod alloc info: %v", alloc)
	}

	var (
		err            error
		containerIPNet *types.IPNetSet
		serviceCIDR    *types.IPNetSet
	)

	serviceCIDR, err = types.ToIPNetSet(alloc.GetBasicInfo().GetServiceCIDR())
	if err != nil {
		return nil, err
	}

	podIP := alloc.GetBasicInfo().GetPodIP()
	subNet := alloc.GetBasicInfo().GetPodCIDR()
	containerIPNet, err = types.BuildIPNet(podIP, subNet)
	if err != nil {
		return nil, err
	}

	containerIfName := alloc.IfName
	if containerIfName == "" {
		containerIfName = args.IfName
	}

	return &types.TeardownCfg{
		ContainerIfName: containerIfName,
		ContainerIPNet:  containerIPNet,
		ServiceCIDR:     serviceCIDR,
	}, nil
}

func GetFileLock(path string) (*filemutex.FileMutex, error) {
	path, _ = filepath.Abs(path)
	lock, err := filemutex.New(path)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("failed to open lock %s: %v", path, err))
	}

	err = wait.PollImmediate(200*time.Millisecond, 10*time.Second, func() (done bool, err error) {
		if err := lock.Lock(); err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}

	return lock, nil
}
