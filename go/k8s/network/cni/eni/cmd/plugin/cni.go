package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"runtime"
	"time"

	"k8s-lx1036/k8s/network/cni/eni/cmd/plugin/driver"
	"k8s-lx1036/k8s/network/cni/eni/pkg/link"
	"k8s-lx1036/k8s/network/cni/eni/pkg/rpc"
	eniTypes "k8s-lx1036/k8s/network/cni/eni/pkg/types"

	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	cniversion "github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

const (
	defaultMTU = 1500

	defaultCniTimeout = 120 * time.Second
	defaultSocketPath = "/var/run/eni/eni.socket"

	defaultEventTimeout = 10 * time.Second

	defaultVethForENI = "veth1"
	defaultVethPrefix = "cali"
)

var (
	veth, rawNIC, ipvlan driver.NetnsDriver

	LinkIP   = net.IPv4(169, 254, 1, 1)
	LinkIPv6 = net.ParseIP("fe80::1")
)

// NetConf is the cni network config
type NetConf struct {
	types.NetConf

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
	types.CommonArgs
	IP                         net.IP
	K8S_POD_NAME               types.UnmarshallableString // nolint
	K8S_POD_NAMESPACE          types.UnmarshallableString // nolint
	K8S_POD_INFRA_CONTAINER_ID types.UnmarshallableString // nolint
}

func init() {
	// This ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, GetSpecVersionSupported(), "")
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
	if err = types.LoadArgs(args.Args, &k8sConfig); err != nil {
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

func initDrivers(ipv4, ipv6 bool) {
	veth = driver.NewVETHDriver(ipv4, ipv6)
	//ipvlan = driver.NewIPVlanDriver(ipv4, ipv6)
	rawNIC = driver.NewRawNICDriver(ipv4, ipv6)
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

	ipv4, ipv6 := allocateIPReply.IPv4, allocateIPReply.IPv6
	initDrivers(ipv4, ipv6)
	hostIPSet, err := driver.GetHostIP(ipv4, ipv6)
	hostVETHName, _ := link.VethNameForPod(string(k8sConfig.K8S_POD_NAME), string(k8sConfig.K8S_POD_NAMESPACE), defaultVethPrefix)

	var containerIPNet *eniTypes.IPNetSet
	var gatewayIPSet *eniTypes.IPSet
	switch allocateIPReply.IPType {
	case rpc.IPType_TypeVPCENI:
		if allocateIPReply.GetENIInfo() == nil || allocateIPReply.GetBasicInfo() == nil ||
			allocateIPReply.GetPod() == nil {
			return fmt.Errorf("vpcEni ip result is empty: %v", allocateIPReply)
		}

		serviceCIDR := allocateIPReply.GetBasicInfo().GetServiceCIDR()
		var svc *eniTypes.IPNetSet
		svc, err = eniTypes.ToIPNetSet(serviceCIDR)
		if err != nil {
			return err
		}

		var extraRoutes []types.Route
		if ipv4 {
			extraRoutes = append(extraRoutes, types.Route{Dst: *svc.IPv4, GW: LinkIP})
		}
		if ipv6 {
			extraRoutes = append(extraRoutes, types.Route{Dst: *svc.IPv6, GW: LinkIPv6})
		}
		for _, v := range conf.HostStackCIDRs {
			_, cidr, err := net.ParseCIDR(v)
			if err != nil {
				return fmt.Errorf("host_stack_cidrs(%s) is invaild: %v", v, err)

			}
			r := types.Route{
				Dst: *cidr,
			}

			if eniTypes.IPv6(cidr.IP) {
				r.GW = LinkIPv6
			} else {
				r.GW = LinkIP
			}
			extraRoutes = append(extraRoutes, r)
		}

		podIP := allocateIPReply.GetBasicInfo().GetPodIP()
		gatewayIP := allocateIPReply.GetBasicInfo().GetGatewayIP()
		eniMAC := allocateIPReply.GetENIInfo().GetMAC()
		containerIPNet, err = eniTypes.BuildIPNet(podIP, &rpc.IPSet{IPv4: "0.0.0.0/32", IPv6: "::/128"})
		if err != nil {
			return err
		}
		gatewayIPSet, err = eniTypes.ToIPSet(gatewayIP)
		if err != nil {
			return err
		}
		var deviceID int32
		deviceID, err = link.GetDeviceNumber(eniMAC)
		if err != nil {
			return err
		}
		ingress := allocateIPReply.GetPod().GetIngress()
		egress := allocateIPReply.GetPod().GetEgress()

		// TODO: file lock

		setupCfg := &driver.SetupConfig{
			HostVETHName:    hostVETHName,
			ContainerIfName: defaultVethForENI,
			ContainerIPNet:  containerIPNet,
			GatewayIP:       gatewayIPSet,
			MTU:             conf.MTU,
			ENIIndex:        int(deviceID),
			Ingress:         ingress,
			Egress:          egress,
			ExtraRoutes:     extraRoutes,
			HostIPSet:       hostIPSet,
		}

		err = veth.Setup(setupCfg, cniNetns)
		if err != nil {
			return fmt.Errorf("setup veth network for eni failed: %v", err)
		}
		defer func() {
			if err != nil {
				if e := veth.Teardown(&driver.TeardownCfg{
					HostVETHName:    hostVETHName,
					ContainerIfName: args.IfName,
				}, cniNetns); e != nil {
					err = errors.Wrapf(err, "tear down veth network for eni failed: %v", e)
				}
			}
		}()

		setupCfg.ContainerIfName = args.IfName
		err = rawNIC.Setup(setupCfg, cniNetns)
		if err != nil {
			return fmt.Errorf("setup network for vpc eni failed: %v", err)
		}

	case rpc.IPType_TypeENIMultiIP:

	case rpc.IPType_TypeVPCIP:

	default:
		return fmt.Errorf("not support this network type")
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

	return types.PrintResult(result, confVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	panic("not implemented")
}

func cmdDel(args *skel.CmdArgs) error {
	panic("not implemented")
}
