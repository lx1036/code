package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/ipam"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/storage"
	"k8s.io/klog/v2"
	"os"
	"sync"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/rpc"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"
)

// TODO: 把 map[PodInfo]ENI 信息存入 pods boltdb 中！！！

const (
	daemonModeVPC        = "VPC"
	daemonModeENIMultiIP = "ENIMultiIP"
	daemonModeENIOnly    = "ENIOnly"

	cniDefaultPath = "/opt/cni/bin"

	dbPath = "/var/lib/cni/pods.db"
	dbName = "pods"
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

	k8sService *K8sService

	cniBinPath string

	eniIPResMgr ResourceManager
	enableTrunk bool

	ipFamily *types.IPFamily

	pendingPods sync.Map // 并发安全的 map
	storage     *storage.DiskStorage
	ipam        ipam.API
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

		k8sService: newK8sServiceOrDie(kubeconfig, daemonMode),
	}

	daemonConfig, err := GetDaemonConfig(configFilePath)
	if err != nil {
		return nil, err
	}

	server.storage, err = storage.NewDiskStorage(dbName, dbPath, json.Marshal, func(bytes []byte) (interface{}, error) {
		var podResource types.PodResources
		if err := json.Unmarshal(bytes, &podResource); err != nil {
			return nil, err
		}
		return podResource, nil
	})
	if err != nil {
		return nil, err
	}
	server.restoreLocalENI()
	eniList, err := server.storage.List()
	if err != nil {
		return nil, err
	}
	localResource := make(map[string]types.PodResource)
	for _, obj := range eniList {
		podResource := obj.(types.PodResources)
		for _, resource := range podResource.Resources {
			localResource[resource.ID] = types.PodResource{
				Resource: resource,
				PodInfo:  podResource.PodInfo,
			}
		}
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
		server.eniIPResMgr, err = newENIIPResourceManager(poolConfig, ecs, server.k8s, localResource)
		if err != nil {
			return nil, err
		}

	}

	return server, nil
}

// restore if local is empty
func (server *EniBackendServer) restoreLocalENI() error {
	items, err := server.storage.List()
	if err != nil {
		return err
	}
	if len(items) != 0 {
		klog.Warningf(fmt.Sprintf("local db is not empty, skip restore"))
		return nil
	}

	eniList, err := server.ipam.GetAttachedENIs(context.TODO(), false)
	if err != nil {
		return err
	}
	ipENIMap := make(map[string]*types.ENI)
	for _, eni := range eniList {
		ipENIMap[eni.PrimaryIP.IPv4.String()] = eni
	}

	podsInfo, err := server.k8sService.GetLocalPods()
	if err != nil {
		return err
	}
	for _, podInfo := range podsInfo {
		if eni, ok := ipENIMap[podInfo.PodIPs.IPv4.String()]; ok {
			server.storage.Put(podInfoKey(podInfo.Namespace, podInfo.Name), types.PodResources{
				PodInfo:   podInfo,
				Resources: eni.ToResItems(),
			})
		}
	}

}

func (server *EniBackendServer) AllocateIP(ctx context.Context, request *rpc.AllocateIPRequest) (*rpc.AllocateIPReply, error) {

	server.RLock()
	defer server.RUnlock()

	// 0. Get pod Info
	podInfo, err := server.k8sService.GetPod(request.K8SPodNamespace, request.K8SPodName)

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
	server.RLock()
	defer server.RUnlock()

	// 0. Get pod Info
	podInfo, err := server.k8sService.GetPod(request.K8SPodNamespace, request.K8SPodName)

	releaseReply := &rpc.ReleaseIPReply{
		Success: true,
		IPv4:    true,
	}

	podKey := fmt.Sprintf("%s/%s")
	oldRes, err := server.getPodResource(podInfoKey(request.K8SPodNamespace, request.K8SPodName))
	if err != nil {
		return nil, err
	}

}

func (server *EniBackendServer) GetIPInfo(ctx context.Context, request *rpc.GetInfoRequest) (*rpc.GetInfoReply, error) {
	// 1. Get pod Info
	podinfo, err := server.k8sService.GetPod(request.K8SPodNamespace, request.K8SPodName)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("get pod info request %+v, err:%v", request.String(), err))
	}

	// 1. Init Context
	networkContext := &networkContext{
		Context:    ctx,
		resources:  []types.ResourceItem{},
		pod:        podinfo,
		k8sService: server.k8sService,
	}
	getIPInfoResult := &rpc.GetInfoReply{IPv4: server.ipFamily.IPv4, IPv6: server.ipFamily.IPv6}

	server.RLock()
	podRes, err := server.getPodResource(podinfo)
	server.RUnlock()
	if err != nil {
		return getIPInfoResult, err
	}

	var netConf []*rpc.NetConf
	// 2. return network info for pod
	switch podinfo.PodNetworkType {
	case podNetworkTypeENIMultiIP:
		getIPInfoResult.IPType = rpc.IPType_TypeENIMultiIP
		resItems := podRes.GetResourceItemByType(types.ResourceTypeENIIP)
		if len(resItems) > 0 {
			// only have one
			res, err := server.eniIPResMgr.Stat(networkContext, resItems[0].ID)
			if err == nil {
				eniIP := res.(*types.ENIIP)
				netConf = append(netConf, &rpc.NetConf{
					BasicInfo: &rpc.BasicInfo{
						PodIP:       eniIP.IPSet.ToRPC(),
						PodCIDR:     eniIP.ENI.VSwitchCIDR.ToRPC(),
						GatewayIP:   eniIP.ENI.GatewayIP.ToRPC(),
						ServiceCIDR: server.k8sService.GetServiceCIDR().ToRPC(),
					},
					ENIInfo: &rpc.ENIInfo{
						MAC:   eniIP.ENI.MAC,
						Trunk: false,
					},
					Pod: &rpc.Pod{
						Ingress: podinfo.TcIngress,
						Egress:  podinfo.TcEgress,
					},
					IfName:      "",
					ExtraRoutes: nil,
				})
			}
		}
		if err = defaultForNetConf(netConfs); err != nil {
			return getIPInfoResult, err
		}

	}

	getIPInfoResult.NetConfs = netConf
	getIPInfoResult.EnableTrunking = server.enableTrunk

	return getIPInfoResult, nil
}

func (server *EniBackendServer) RecordEvent(ctx context.Context, request *rpc.EventRequest) (*rpc.EventReply, error) {
	panic("implement me")
}

func (server *EniBackendServer) getPodResource(key string) (types.PodResources, error) {

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
