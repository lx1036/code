package watchers

import (
	"fmt"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/service"
	"time"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s"

	"k8s.io/klog/v2"
)

const (
	cacheSyncTimeout = 3 * time.Minute

	metricCNP            = "CiliumNetworkPolicy"
	metricCCNP           = "CiliumClusterwideNetworkPolicy"
	metricEndpoint       = "Endpoint"
	metricEndpointSlice  = "EndpointSlice"
	metricKNP            = "NetworkPolicy"
	metricNS             = "Namespace"
	metricCiliumNode     = "CiliumNode"
	metricCiliumEndpoint = "CiliumEndpoint"
	metricPod            = "Pod"
	metricNode           = "Node"
	metricService        = "Service"
	metricCreate         = "create"
	metricDelete         = "delete"
	metricUpdate         = "update"
)

type K8sWatcher struct {

	// K8sSvcCache is a cache of all Kubernetes services and endpoints
	K8sSvcCache k8s.ServiceCache

	serviceBPFManager *service.ServiceBPFManager
}

func NewK8sWatcher(
	serviceBPFManager *service.ServiceBPFManager,
) *K8sWatcher {

	return &K8sWatcher{
		K8sSvcCache: k8s.NewServiceCache(datapath.LocalNodeAddressing()),

		serviceBPFManager: serviceBPFManager,
	}
}

func (k *K8sWatcher) InitK8sSubsystem() <-chan struct{} {
	if err := k.EnableK8sWatcher(); err != nil {
		klog.Fatal("Unable to establish connection to Kubernetes apiserver")
	}

	cachesSynced := make(chan struct{})

	go func() {
		// wait for cache sync data from api-server
		klog.Info("Waiting until all pre-existing resources related to policy have been received")

		close(cachesSynced)
	}()

	go func() {
		select {
		case <-cachesSynced:
			klog.Info("All pre-existing resources related to policy have been received; continuing")
		case <-time.After(cacheSyncTimeout):
			klog.Fatalf("Timed out waiting for pre-existing resources related to policy to be received; exiting")
		}
	}()

	return cachesSynced
}

// EnableK8sWatcher watch k8s service/endpoint/networkpolicy
func (k *K8sWatcher) EnableK8sWatcher() error {

	// watch kubernetes services
	k.watchK8sService(k8s.WatcherCli(), swgSvcs)

	// watch kubernetes either "Endpoints" or "EndpointSlice"
	switch {
	case k8s.SupportsEndpointSlice():
		connected := k.watchEndpointSlices(k8s.WatcherCli(), swgEps)
		// the cluster has endpoint slices so we should not check for v1.Endpoints
		if connected {
			break
		}
		fallthrough
	default:
		k.watchEndpoints(k8s.WatcherCli(), swgEps)
	}

}

func (k *K8sWatcher) RunK8sServiceHandler() {
	go k.k8sServiceHandler()
}

func (k *K8sWatcher) k8sServiceHandler() {

	eventHandler := func(event k8s.ServiceEvent) {
		defer event.SWG.Done()

		svc := event.Service

		switch event.Action {
		case k8s.UpdateService:
			if err := k.addK8sSVCs(event.ID, event.OldService, svc, event.Endpoints); err != nil {
				klog.Errorf(fmt.Sprintf("Unable to add/update service to implement k8s event err:%v", err))
			}

			if !svc.IsExternal() {
				return
			}

			// TODO: network policy
		case k8s.DeleteService:
			if err := k.delK8sSVCs(event.ID, event.Service, event.Endpoints); err != nil {
				klog.Errorf(fmt.Sprintf("Unable to delete service to implement k8s event err:%v", err))
			}

			if !svc.IsExternal() {
				return
			}
			// TODO: network policy
		}
	}

	for {
		event, ok := <-k.K8sSvcCache.Events
		if !ok {
			return
		}
		eventHandler(event)
	}
}

func (k *K8sWatcher) addK8sSVCs(svcID k8s.ServiceID, oldSvc, svc *k8s.Service, endpoints *k8s.Endpoints) error {

	// Headless services do not need any datapath implementation
	if svc.IsHeadless {
		return nil
	}

	svcs := datapathSVCs(svc, endpoints)
	svcMap := hashSVCMap(svcs)

	if oldSvc != nil {
		// If we have oldService then we need to detect which frontends
		// are no longer in the updated service and delete them in the datapath.

		oldSVCs := datapathSVCs(oldSvc, endpoints)
		oldSVCMap := hashSVCMap(oldSVCs)

		for svcHash, oldSvc := range oldSVCMap {
			if _, ok := svcMap[svcHash]; !ok {
				if found, err := k.serviceBPFManager.DeleteService(oldSvc); err != nil {
					scopedLog.WithError(err).WithField(logfields.Object, logfields.Repr(oldSvc)).
						Warn("Error deleting service by frontend")
				} else if !found {
					scopedLog.WithField(logfields.Object, logfields.Repr(oldSvc)).Warn("service not found")
				} else {
					scopedLog.Debugf("# cilium lb delete-service %s %d 0", oldSvc.IP, oldSvc.Port)
				}
			}
		}
	}

	for _, dpSvc := range svcs {
		if _, _, err := k.serviceBPFManager.UpdateOrInsertService(dpSvc.Frontend, dpSvc.Backends, dpSvc.Type,
			dpSvc.TrafficPolicy,
			dpSvc.SessionAffinity, dpSvc.SessionAffinityTimeoutSec,
			dpSvc.HealthCheckNodePort,
			svcID.Name, svcID.Namespace); err != nil {
			scopedLog.WithError(err).Error("Error while inserting service in LB map")
		}
	}
	return nil
}

func (k *K8sWatcher) delK8sSVCs(svc k8s.ServiceID, svcInfo *k8s.Service, se *k8s.Endpoints) error {

}
