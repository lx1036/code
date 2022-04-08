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

	}

	return server, nil
}

func (server *EniBackendServer) AllocateIP(ctx context.Context, request *rpc.AllocateIPRequest) (*rpc.AllocateIPReply, error) {

	server.RLock()
	defer server.RUnlock()

	// 0. Get pod Info
	podinfo, err := server.k8s.GetPod(r.K8SPodNamespace, r.K8SPodName)

	// 1. Init Context
	allocIPReply := &rpc.AllocateIPReply{IPv4: server.ipFamily.IPv4, IPv6: server.ipFamily.IPv6}

	// 3. Allocate network resource for pod
	switch podInfo.PodNetworkType {

	case podNetworkTypeVPCENI:
		var eni *types.ENI
		eni, err = server.allocateENI(networkContext, &oldRes)
		if err != nil {
			return nil, fmt.Errorf("error get allocated vpc ENI ip for: %+v, result: %+v", podinfo, err)
		}

		allocIPReply.IPType = rpc.IPType_TypeVPCENI
		allocIPReply.Success = true
		allocIPReply.BasicInfo = &rpc.BasicInfo{
			PodIP:       eni.PrimaryIP.ToRPC(),
			PodCIDR:     eni.VSwitchCIDR.ToRPC(),
			GatewayIP:   eni.GatewayIP.ToRPC(),
			ServiceCIDR: server.k8s.GetServiceCIDR().ToRPC(),
		}
		allocIPReply.ENIInfo = &rpc.ENIInfo{
			MAC:   eni.MAC,
			Trunk: podinfo.PodENI && server.enableTrunk && eni.Trunk,
		}
		allocIPReply.Pod = &rpc.Pod{
			Ingress: podinfo.TcIngress,
			Egress:  podinfo.TcEgress,
		}
	default:
		return nil, fmt.Errorf("not support pod network type")
	}

	// 4. grpc connection
	if ctx.Err() != nil {
		err = ctx.Err()
		return nil, errors.Wrapf(err, "error on grpc connection")
	}

	// 5. return allocate result
	return allocIPReply, err
}

func (server *EniBackendServer) allocateENI() (*types.ENI, error) {

	res, err := server.eniResMgr.Allocate(ctx, oldENIID)
	if err != nil {
		return nil, err
	}

	return res.(*types.ENI), nil
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
