package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"runtime"
	"time"

	"k8s-lx1036/k8s/network/cni/eni/pkg/link"
	"k8s-lx1036/k8s/network/cni/eni/plugin/driver"
	"k8s-lx1036/k8s/network/cni/eni/plugin/version"
	"k8s-lx1036/k8s/network/cni/eni/rpc"
	eniTypes "k8s-lx1036/k8s/network/cni/eni/types"

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
)

// NetConf is the cni network config
type NetConf struct {
	// CNIVersion is the plugin version
	CNIVersion string `json:"cniVersion,omitempty"`

	// Name is the plugin name
	Name string `json:"name"`

	// Type is the plugin type
	Type string `json:"type"`

	// HostVethPrefix is the veth for container prefix on host
	HostVethPrefix string `json:"veth_prefix"`

	// eniIPVirtualType is the ipvlan for container
	ENIIPVirtualType string `json:"eniip_virtual_type"`

	// HostStackCIDRs is a list of CIDRs, all traffic targeting these CIDRs will be redirected to host network stack
	HostStackCIDRs []string `json:"host_stack_cidrs"`

	// MTU is container and ENI network interface MTU
	MTU int `json:"mtu"`

	// Debug
	Debug bool `json:"debug"`
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
	runtime.LockOSThread()
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.GetSpecVersionSupported(), "")
}

func parseCmdArgs(args *skel.CmdArgs) (string, ns.NetNS, *NetConf, *K8SArgs, error) {

	versionDecoder := &cniversion.ConfigDecoder{}
	confVersion, err := versionDecoder.Decode(args.StdinData)
	if err != nil {
		return "", nil, nil, nil, err
	}
	netNS, err := ns.GetNS(args.Netns)
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
	ipvlan = driver.NewIPVlanDriver(ipv4, ipv6)
	rawNIC = driver.NewRawNICDriver(ipv4, ipv6)
}

func cmdAdd(args *skel.CmdArgs) error {

	confVersion, cniNetns, conf, k8sConfig, err := parseCmdArgs(args)
	if err != nil {
		return err
	}
	defer cniNetns.Close()

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
		return fmt.Errorf("cmdAdd: error alloc ip for pod %s, %w", driver.PodInfoKey(string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME)), err)
	}
	if !allocateIPReply.Success {
		return fmt.Errorf("cmdAdd: error alloc ip for pod %s", driver.PodInfoKey(string(k8sConfig.K8S_POD_NAMESPACE), string(k8sConfig.K8S_POD_NAME)))
	}
	klog.Infof("%#v", allocateIPReply)
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

}

func cmdDel(args *skel.CmdArgs) error {

}
