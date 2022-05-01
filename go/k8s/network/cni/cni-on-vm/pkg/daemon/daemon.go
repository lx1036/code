package daemon

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"sync"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/rpc"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/types"
)

const (
	daemonModeVPC        = "VPC"
	daemonModeENIMultiIP = "ENIMultiIP"
	daemonModeENIOnly    = "ENIOnly"

	cniDefaultPath = "/opt/cni/bin"
)

// ResourceManager Allocate/Release/Pool/Stick/GC pod resource
// managed pod and resource relationship
type ResourceManager interface {
	Allocate(context *networkContext, prefer string) (types.NetworkResource, error)
	Release(context *networkContext, resItem types.ResourceItem) error
	GarbageCollection(inUseResSet map[string]types.ResourceItem, expireResSet map[string]types.ResourceItem) error
	Stat(context *networkContext, resID string) (types.NetworkResource, error)
	//tracing.ResourceMappingHandler
}

type EniBackendServer struct {
	rpc.UnimplementedEniBackendServer

	sync.RWMutex

	daemonMode     string
	configFilePath string
	kubeConfig     string
	master         string

	cniBinPath string

	eniIPResMgr ResourceManager
	enableTrunk bool

	ipFamily *types.IPFamily

	pendingPods sync.Map // 并发安全的 map

}

func newEniBackendServer(daemonMode, configFilePath, kubeconfig string) (rpc.EniBackendServer, error) {
	cniBinPath := os.Getenv("CNI_PATH")
	if cniBinPath == "" {
		cniBinPath = cniDefaultPath
	}
	server := &EniBackendServer{
		configFilePath: configFilePath,
		kubeConfig:     kubeconfig,
		//master:         master,
		pendingPods: sync.Map{},
		cniBinPath:  cniBinPath,
	}

	daemonConfig, err := GetDaemonConfig(configFilePath)
	if err != nil {
		return nil, err
	}

	server.enableTrunk = daemonConfig.EnableENITrunking

	switch daemonMode {
	case daemonModeENIOnly, daemonModeENIMultiIP, daemonModeVPC:
		server.daemonMode = daemonMode
	default:
		return nil, fmt.Errorf("unsupport daemon mode %s", daemonMode)
	}

	var err error
	switch daemonMode {
	case daemonModeENIMultiIP:
		server.eniIPResMgr, err = newENIIPResourceManager(poolConfig, ecs, server.k8s, localResource[types.ResourceTypeENI])
		if err != nil {
			return nil, err
		}

	}

	return server, nil
}

func (server *EniBackendServer) AllocateIP(ctx context.Context, request *rpc.AllocateIPRequest) (*rpc.AllocateIPReply, error) {

	server.RLock()
	defer server.RUnlock()

	// 0. Get pod Info
	podInfo, err := server.k8s.GetPod(r.K8SPodNamespace, r.K8SPodName)

	// 1. Init Context
	allocIPReply := &rpc.AllocateIPReply{
		IPv4: server.ipFamily.IPv4,
		IPv6: server.ipFamily.IPv6,
	}

	// 3. Allocate network resource for pod
	var netConfs []*rpc.NetConf
	switch podInfo.PodNetworkType {

	case podNetworkTypeENIMultiIP:
		eniIP, err := server.allocateENIMultiIP(networkContext, &oldRes)

		netConfs = append(netConfs, &rpc.NetConf{
			BasicInfo: &rpc.BasicInfo{
				PodIP:       eniIP.IPSet.ToRPC(),
				PodCIDR:     eniIP.ENI.VSwitchCIDR.ToRPC(),
				GatewayIP:   eniIP.ENI.GatewayIP.ToRPC(),
				ServiceCIDR: server.k8s.GetServiceCIDR().ToRPC(),
			},
			ENIInfo: &rpc.ENIInfo{
				MAC:   eniIP.ENI.MAC,
				Trunk: false,
			},
			Pod: &rpc.Pod{
				Ingress: podInfo.TcIngress,
				Egress:  podInfo.TcEgress,
			},
			IfName:       "",
			ExtraRoutes:  nil,
			DefaultRoute: true,
		})
		if err = defaultForNetConf(netConfs); err != nil {
			return nil, err
		}
		allocIPReply.Success = true

	default:
		return nil, fmt.Errorf("not support pod network type")
	}

	// 4. grpc connection
	if ctx.Err() != nil {
		err = ctx.Err()
		return nil, errors.Wrapf(err, "error on grpc connection")
	}

	// 5. return allocate result
	allocIPReply.NetConfs = netConfs
	allocIPReply.EnableTrunking = server.enableTrunk
	return allocIPReply, err
}

func (server *EniBackendServer) allocateENIMultiIP(ctx *networkContext, old *types.PodResources) (*types.ENIIP, error) {
	oldENIIPID := ""

	res, err := server.eniIPResMgr.Allocate(ctx, oldENIIPID)
	if err != nil {
		return nil, err
	}

	return res.(*types.ENIIP), nil
}

func (server *EniBackendServer) ReleaseIP(ctx context.Context, request *rpc.ReleaseIPRequest) (*rpc.ReleaseIPReply, error) {
	panic("implement me")
}

func (server *EniBackendServer) GetIPInfo(ctx context.Context, request *rpc.GetInfoRequest) (*rpc.GetInfoReply, error) {
	panic("implement me")
}

func (server *EniBackendServer) RecordEvent(ctx context.Context, request *rpc.EventRequest) (*rpc.EventReply, error) {
	panic("implement me")
}

func defaultForNetConf(netConf []*rpc.NetConf) error {
	// ignore netConf check
	if len(netConf) == 0 {
		return nil
	}
	defaultRouteSet := false
	defaultIfSet := false
	for i := 0; i < len(netConf); i++ {
		if netConf[i].DefaultRoute && defaultRouteSet {
			return fmt.Errorf("default route is dumplicated")
		}
		defaultRouteSet = defaultRouteSet || netConf[i].DefaultRoute

		if defaultIf(netConf[i].IfName) {
			defaultIfSet = true
		}
	}

	if !defaultIfSet {
		return fmt.Errorf("default interface is not set")
	}

	if !defaultRouteSet {
		for i := 0; i < len(netConf); i++ {
			if netConf[i].IfName == "" || netConf[i].IfName == "eth0" {
				netConf[i].DefaultRoute = true
				break
			}
		}
	}

	return nil
}

func defaultIf(name string) bool {
	if name == "" || name == "eth0" {
		return true
	}
	return false
}
