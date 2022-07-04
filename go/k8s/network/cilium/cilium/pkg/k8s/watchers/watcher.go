package watchers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/cache"
	"net"
	"strconv"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint/endpointmanager"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/service"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"

	k8smetrics "github.com/cilium/cilium/pkg/k8s/metrics"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
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
	K8sSvcCache     k8s.ServiceCache
	endpointManager *endpointmanager.EndpointManager
	podStore        cache.Store

	serviceBPFManager *service.ServiceBPFManager
}

func NewK8sWatcher(
	endpointManager *endpointmanager.EndpointManager,
	serviceBPFManager *service.ServiceBPFManager,
) *K8sWatcher {

	return &K8sWatcher{
		K8sSvcCache: k8s.NewServiceCache(datapath.LocalNodeAddressing()),

		endpointManager:   endpointManager,
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

	wg := &sync.WaitGroup{}

	k.watchK8sNetworkPolicy(k8s.WatcherCli())

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
		k.watchK8sEndpoints(k8s.WatcherCli(), swgEps)
	}

	// watch kubernetes pods
	wg.Add(1)
	go k.watchK8sPod(k8s.WatcherCli())

	// watch kubernetes nodes
	k.watchK8sNode(k8s.WatcherCli())

	// watch kubernetes namespace
	wg.Add(1)
	go k.watchK8sNamespace(k8s.WatcherCli())

	wg.Wait()

	return nil
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
			if err := k.addK8sSvcIntoBPF(event.ID, event.OldService, svc, event.Endpoints); err != nil {
				klog.Errorf(fmt.Sprintf("Unable to add/update service to implement k8s event err:%v", err))
			}

			if !svc.IsExternal() {
				return
			}

			// TODO: network policy
		case k8s.DeleteService:
			if err := k.delK8sSvcFromBPF(event.ID, event.Service, event.Endpoints); err != nil {
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

func (k *K8sWatcher) addK8sSvcIntoBPF(svcID k8s.ServiceID, oldSvc, svc *k8s.Service, endpoints *k8s.Endpoints) error {
	// Headless services do not need any datapath implementation
	if svc.IsHeadless {
		return nil
	}

	scopedLog := log.WithFields(log.Fields{
		logfields.K8sSvcName:   svcID.Name,
		logfields.K8sNamespace: svcID.Namespace,
	})

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

func (k *K8sWatcher) delK8sSvcFromBPF(svcID k8s.ServiceID, svc *k8s.Service, se *k8s.Endpoints) error {
	// Headless services do not need any datapath implementation
	if svc.IsHeadless {
		return nil
	}

	scopedLog := log.WithFields(log.Fields{
		logfields.K8sSvcName:   svcID.Name,
		logfields.K8sNamespace: svcID.Namespace,
	})

	repPorts := svc.UniquePorts()
	var frontends []*loadbalancer.L3n4Addr
	for portName, svcPort := range svc.Ports {
		if !repPorts[svcPort.Port] {
			continue
		}
		repPorts[svcPort.Port] = false

		fe := loadbalancer.NewL3n4Addr(svcPort.Protocol, svc.FrontendIP, svcPort.Port, loadbalancer.ScopeExternal)
		frontends = append(frontends, fe)

		// NodePort
		for _, nodePortFE := range svc.NodePorts[portName] {
			frontends = append(frontends, &nodePortFE.L3n4Addr)
			if svc.TrafficPolicy == loadbalancer.SVCTrafficPolicyLocal {
				cpFE := nodePortFE.L3n4Addr.DeepCopy()
				cpFE.Scope = loadbalancer.ScopeInternal
				frontends = append(frontends, cpFE)
			}
		}

		for _, k8sExternalIP := range svc.K8sExternalIPs {
			frontends = append(frontends, loadbalancer.NewL3n4Addr(svcPort.Protocol, k8sExternalIP, svcPort.Port, loadbalancer.ScopeExternal))
		}

		for _, ip := range svc.LoadBalancerIPs {
			frontends = append(frontends, loadbalancer.NewL3n4Addr(svcPort.Protocol, ip, svcPort.Port, loadbalancer.ScopeExternal))
			if svc.TrafficPolicy == loadbalancer.SVCTrafficPolicyLocal {
				frontends = append(frontends, loadbalancer.NewL3n4Addr(svcPort.Protocol, ip, svcPort.Port, loadbalancer.ScopeInternal))
			}
		}
	}

	for _, fe := range frontends {
		if found, err := k.serviceBPFManager.DeleteService(*fe); err != nil {
			scopedLog.WithError(err).WithField(logfields.Object, logfields.Repr(fe)).
				Warn("Error deleting service by frontend")
		} else if !found {
			scopedLog.WithField(logfields.Object, logfields.Repr(fe)).Warn("service not found")
		} else {
			scopedLog.Debugf("# cilium lb delete-service %s %d 0", fe.IP, fe.Port)
		}
	}

	return nil
}

// datapathSVCs returns all services that should be set in the datapath.
func datapathSVCs(svc *k8s.Service, endpoints *k8s.Endpoints) (svcs []loadbalancer.SVC) {
	uniqPorts := svc.UniquePorts()
	clusterIPPorts := map[loadbalancer.FEPortName]*loadbalancer.L4Addr{}
	for fePortName, fePort := range svc.Ports {
		if !uniqPorts[fePort.Port] {
			continue
		}
		uniqPorts[fePort.Port] = false
		clusterIPPorts[fePortName] = fePort
	}

	// ClusterIP/LoadBalancer/ExternalIP/NodePort
	if svc.FrontendIP != nil {
		dpSVC := genCartesianProduct(svc.FrontendIP, svc.TrafficPolicy, loadbalancer.SVCTypeClusterIP, clusterIPPorts, endpoints)
		svcs = append(svcs, dpSVC...)
	}
	for _, ip := range svc.LoadBalancerIPs {
		dpSVC := genCartesianProduct(ip, svc.TrafficPolicy, loadbalancer.SVCTypeLoadBalancer, clusterIPPorts, endpoints)
		svcs = append(svcs, dpSVC...)
	}
	for _, k8sExternalIP := range svc.K8sExternalIPs {
		dpSVC := genCartesianProduct(k8sExternalIP, svc.TrafficPolicy, loadbalancer.SVCTypeExternalIPs, clusterIPPorts, endpoints)
		svcs = append(svcs, dpSVC...)
	}
	for fePortName := range clusterIPPorts {
		for _, nodePortFE := range svc.NodePorts[fePortName] {
			nodePortPorts := map[loadbalancer.FEPortName]*loadbalancer.L4Addr{
				fePortName: &nodePortFE.L4Addr,
			}
			dpSVC := genCartesianProduct(nodePortFE.IP, svc.TrafficPolicy, loadbalancer.SVCTypeNodePort, nodePortPorts, endpoints)
			svcs = append(svcs, dpSVC...)
		}
	}

	// apply common service properties
	for i := range svcs {
		svcs[i].TrafficPolicy = svc.TrafficPolicy
		svcs[i].HealthCheckNodePort = svc.HealthCheckNodePort
		svcs[i].SessionAffinity = svc.SessionAffinity
		svcs[i].SessionAffinityTimeoutSec = svc.SessionAffinityTimeoutSec
	}

	return svcs
}

func genCartesianProduct(
	fe net.IP,
	svcTrafficPolicy loadbalancer.SVCTrafficPolicy,
	svcType loadbalancer.SVCType,
	ports map[loadbalancer.FEPortName]*loadbalancer.L4Addr,
	bes *k8s.Endpoints,
) []loadbalancer.SVC {
	var svcSize int

	// For externalTrafficPolicy=Local we add both external and internal
	// scoped frontends, hence twice the size for only this case.
	if svcTrafficPolicy == loadbalancer.SVCTrafficPolicyLocal &&
		(svcType == loadbalancer.SVCTypeLoadBalancer || svcType == loadbalancer.SVCTypeNodePort) {
		svcSize = len(ports) * 2
	} else {
		svcSize = len(ports)
	}

	svcs := make([]loadbalancer.SVC, 0, svcSize)
	for fePortName, fePort := range ports {
		var besValues []loadbalancer.Backend
		for ip, backend := range bes.Backends {
			if backendPort := backend.Ports[string(fePortName)]; backendPort != nil {
				besValues = append(besValues, loadbalancer.Backend{
					NodeName: backend.NodeName,
					L3n4Addr: loadbalancer.L3n4Addr{
						IP: net.ParseIP(ip), L4Addr: *backendPort,
					},
				})
			}
		}

		// External scoped entry.
		svcs = append(svcs,
			loadbalancer.SVC{
				Frontend: loadbalancer.L3n4AddrID{
					L3n4Addr: loadbalancer.L3n4Addr{
						IP: fe,
						L4Addr: loadbalancer.L4Addr{
							Protocol: fePort.Protocol,
							Port:     fePort.Port,
						},
						Scope: loadbalancer.ScopeExternal,
					},
					ID: loadbalancer.ID(0),
				},
				Backends: besValues,
				Type:     svcType,
			})

		// Internal scoped entry only for Local traffic policy.
		if svcSize > len(ports) {
			svcs = append(svcs,
				loadbalancer.SVC{
					Frontend: loadbalancer.L3n4AddrID{
						L3n4Addr: loadbalancer.L3n4Addr{
							IP: fe,
							L4Addr: loadbalancer.L4Addr{
								Protocol: fePort.Protocol,
								Port:     fePort.Port,
							},
							Scope: loadbalancer.ScopeInternal,
						},
						ID: loadbalancer.ID(0),
					},
					Backends: besValues,
					Type:     svcType,
				})
		}
	}

	return svcs
}

// hashSVCMap returns a mapping of all frontend's hash to the its corresponded
// value.
func hashSVCMap(svcs []loadbalancer.SVC) map[string]loadbalancer.L3n4Addr {
	m := map[string]loadbalancer.L3n4Addr{}
	for _, svc := range svcs {
		m[svc.Frontend.L3n4Addr.Hash()] = svc.Frontend.L3n4Addr
	}
	return m
}

// K8sEventProcessed is called to do metrics accounting for each processed
// Kubernetes event
func (k *K8sWatcher) K8sEventProcessed(scope string, action string, status bool) {
	result := "success"
	if status == false {
		result = "failed"
	}

	metrics.KubernetesEventProcessed.WithLabelValues(scope, action, result).Inc()
}

// K8sEventReceived does metric accounting for each received Kubernetes event
func (k *K8sWatcher) K8sEventReceived(scope string, action string, valid, equal bool) {
	metrics.EventTSK8s.SetToCurrentTime()
	k8smetrics.LastInteraction.Reset()

	metrics.KubernetesEventReceived.WithLabelValues(scope, action, strconv.FormatBool(valid), strconv.FormatBool(equal)).Inc()
}
