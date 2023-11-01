package k8s

import (
	"fmt"
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/logging/logfields"
	log "github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"net"
	"strings"
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

// CacheAction is the type of action that was performed on the cache
type CacheAction int

const (
	// UpdateService reflects that the service was updated or added
	UpdateService CacheAction = iota

	// DeleteService reflects that the service was deleted
	DeleteService
)

// ServiceID identifies the Kubernetes service
type ServiceID struct {
	Name      string `json:"serviceName,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// String returns the string representation of a service ID
func (s ServiceID) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}

// ParseServiceID parses a Kubernetes service and returns the ServiceID
func ParseServiceID(svc *corev1.Service) ServiceID {
	return ServiceID{
		Name:      svc.ObjectMeta.Name,
		Namespace: svc.ObjectMeta.Namespace,
	}
}

type ServiceEvent struct {
	// ID is the identified of the service
	ID ServiceID

	Action CacheAction

	// Service is the service structure
	Service *Service

	// OldService is the service structure
	OldService *Service

	// Endpoints is the endpoints structured correlated with the service
	Endpoints *Endpoints
}

// Service is an abstraction for a k8s service that is composed by the frontend IP
// address (FEIP) and the map of the frontend ports (Ports).
// +k8s:deepcopy-gen=true
type Service struct {
	FrontendIP net.IP
	IsHeadless bool

	Labels   map[string]string
	Selector map[string]string

	// HealthCheckNodePort defines on which port the node runs a HTTP health
	// check server which may be used by external loadbalancers to determine
	// if a node has local backends. This will only have effect if both
	// LoadBalancerIPs is not empty and TrafficPolicy is SVCTrafficPolicyLocal.
	HealthCheckNodePort uint16
	Ports               map[loadbalancer.FEPortName]*loadbalancer.L4Addr
	// NodePorts stores mapping for port name => NodePort frontend addr string =>
	// NodePort fronted addr. The string addr => addr indirection is to avoid
	// storing duplicates.
	NodePorts map[loadbalancer.FEPortName]map[string]*loadbalancer.L3n4AddrID
	// K8sExternalIPs stores mapping of the endpoint in a string format to the
	// externalIP in net.IP format.
	K8sExternalIPs map[string]net.IP
	// LoadBalancerIPs stores LB IPs assigned to the service (string(IP) => IP).
	LoadBalancerIPs map[string]net.IP

	// TrafficPolicy controls how backends are selected. If set to "Local", only
	// node-local backends are chosen
	TrafficPolicy loadbalancer.SVCTrafficPolicy
	// SessionAffinity denotes whether service has the clientIP session affinity
	SessionAffinity bool
	// SessionAffinityTimeoutSeconds denotes session affinity timeout
	SessionAffinityTimeoutSec uint32

	// IncludeExternal is true when external endpoints from other clusters
	// should be included
	IncludeExternal bool
	// Shared is true when the service should be exposed/shared to other clusters
	Shared bool
}

func NewService(ip net.IP, externalIPs []string, loadBalancerIPs []string,
	headless bool, trafficPolicy loadbalancer.SVCTrafficPolicy,
	healthCheckNodePort uint16, labels, selector map[string]string) *Service {

	var k8sExternalIPs map[string]net.IP
	var k8sLoadBalancerIPs map[string]net.IP

	k8sExternalIPs = parseIPs(externalIPs)
	k8sLoadBalancerIPs = parseIPs(loadBalancerIPs)

	return &Service{
		FrontendIP:          ip,
		IsHeadless:          headless,
		TrafficPolicy:       trafficPolicy,
		HealthCheckNodePort: healthCheckNodePort,

		Ports:           map[loadbalancer.FEPortName]*loadbalancer.L4Addr{},
		NodePorts:       map[loadbalancer.FEPortName]map[string]*loadbalancer.L3n4AddrID{},
		K8sExternalIPs:  k8sExternalIPs,
		LoadBalancerIPs: k8sLoadBalancerIPs,

		Labels:   labels,
		Selector: selector,
	}
}

func (s *Service) UniquePorts() map[uint16]bool {
	// We are not discriminating the different L4 protocols on the same L4
	// port so we create the number of unique sets of service IP + service
	// port.
	uniqPorts := map[uint16]bool{}
	for _, p := range s.Ports {
		uniqPorts[p.Port] = true
	}
	return uniqPorts
}

// IsExternal returns true if the service is expected to serve out-of-cluster endpoints:
func (s Service) IsExternal() bool {
	return len(s.Selector) == 0
}

func parseIPs(externalIPs []string) map[string]net.IP {
	m := map[string]net.IP{}
	for _, externalIP := range externalIPs {
		ip := net.ParseIP(externalIP)
		if ip != nil {
			m[externalIP] = ip
		}
	}
	return m
}

func ParseService(k8sSvc *corev1.Service, nodeAddressing datapath.Datapath) (ServiceID, *Service) {
	svcID := ParseServiceID(k8sSvc)

	switch k8sSvc.Spec.Type {
	case corev1.ServiceTypeClusterIP, corev1.ServiceTypeNodePort, corev1.ServiceTypeLoadBalancer:
		break

	case corev1.ServiceTypeExternalName:
		// External-name services must be ignored
		return ServiceID{}, nil

	default:
		klog.Warning("Ignoring k8s service: unsupported type")
		return ServiceID{}, nil
	}

	if k8sSvc.Spec.ClusterIP == "" && len(k8sSvc.Spec.ExternalIPs) == 0 {
		return ServiceID{}, nil
	}

	clusterIP := net.ParseIP(k8sSvc.Spec.ClusterIP)
	headless := false
	if strings.ToLower(k8sSvc.Spec.ClusterIP) == "none" {
		headless = true
	}

	// external traffic policy
	var trafficPolicy loadbalancer.SVCTrafficPolicy
	switch k8sSvc.Spec.ExternalTrafficPolicy {
	case corev1.ServiceExternalTrafficPolicyTypeLocal:
		trafficPolicy = loadbalancer.SVCTrafficPolicyLocal
	default:
		trafficPolicy = loadbalancer.SVCTrafficPolicyCluster
	}

	var loadBalancerIPs []string
	for _, ip := range k8sSvc.Status.LoadBalancer.Ingress {
		if ip.IP != "" {
			loadBalancerIPs = append(loadBalancerIPs, ip.IP)
		}
	}

	svcInfo := NewService(clusterIP, k8sSvc.Spec.ExternalIPs, loadBalancerIPs, headless,
		trafficPolicy, uint16(k8sSvc.Spec.HealthCheckNodePort), k8sSvc.Labels, k8sSvc.Spec.Selector)
	svcInfo.IncludeExternal = getAnnotationIncludeExternal(k8sSvc)
	svcInfo.Shared = getAnnotationShared(k8sSvc)
	// session affinity
	if k8sSvc.Spec.SessionAffinity == corev1.ServiceAffinityClientIP {
		svcInfo.SessionAffinity = true
		if cfg := k8sSvc.Spec.SessionAffinityConfig; cfg != nil && cfg.ClientIP != nil && cfg.ClientIP.TimeoutSeconds != nil {
			svcInfo.SessionAffinityTimeoutSec = uint32(*cfg.ClientIP.TimeoutSeconds)
		}
		if svcInfo.SessionAffinityTimeoutSec == 0 {
			svcInfo.SessionAffinityTimeoutSec = uint32(corev1.DefaultClientIPServiceAffinitySeconds)
		}
	}

	// ports/nodePorts
	for _, port := range k8sSvc.Spec.Ports {
		p := loadbalancer.NewL4Addr(loadbalancer.L4Type(port.Protocol), uint16(port.Port))
		portName := loadbalancer.FEPortName(port.Name)
		if _, ok := svcInfo.Ports[portName]; !ok {
			svcInfo.Ports[portName] = p
		}

		if k8sSvc.Spec.Type == corev1.ServiceTypeNodePort || k8sSvc.Spec.Type == corev1.ServiceTypeLoadBalancer {
			if nodeAddressing != nil { // nodeAddressing 为了获取宿主机 nodeIP 等信息
				if _, ok := svcInfo.NodePorts[portName]; !ok {
					svcInfo.NodePorts[portName] = make(map[string]*loadbalancer.L3n4AddrID)
				}

				proto := loadbalancer.L4Type(port.Protocol)
				port := uint16(port.NodePort)
				id := loadbalancer.ID(0) // will be allocated by k8s_watcher
				if clusterIP != nil && !strings.Contains(k8sSvc.Spec.ClusterIP, ":") {
					for _, ip := range nodeAddressing.IPv4().LoadBalancerNodeAddresses() {
						nodePortFE := loadbalancer.NewL3n4AddrID(proto, ip, port, loadbalancer.ScopeExternal, id)
						svcInfo.NodePorts[portName][nodePortFE.String()] = nodePortFE
					}
				}
			}
		}
	}

	return svcID, svcInfo
}

func getAnnotationIncludeExternal(svc *corev1.Service) bool {
	if value, ok := svc.ObjectMeta.Annotations[GlobalService]; ok {
		return strings.ToLower(value) == "true"
	}

	return false
}

func getAnnotationShared(svc *corev1.Service) bool {
	if value, ok := svc.ObjectMeta.Annotations[SharedService]; ok {
		return strings.ToLower(value) == "true"
	}

	return getAnnotationIncludeExternal(svc)
}

type ServiceCache struct {
	mutex sync.RWMutex

	Events chan ServiceEvent

	nodeAddressing datapath.Datapath

	services map[ServiceID]*Service

	// endpoints maps a service to a map of endpointSlices. In case the cluster
	// is still using the v1.Endpoints, the key used in the internal map of
	// endpointSlices is the v1.Endpoint name.
	endpoints map[ServiceID]*endpointSlices
	// externalEndpoints is a list of additional service backends derived from source other than the local cluster
	externalEndpoints map[ServiceID]externalEndpoints
}

func NewServiceCache(nodeAddressing datapath.Datapath) ServiceCache {
	return ServiceCache{

		nodeAddressing: nodeAddressing,
		Events:         make(chan ServiceEvent, option.Config.K8sServiceCacheSize),
	}
}

func (s *ServiceCache) UpdateService(k8sSvc *corev1.Service, swg *lock.StoppableWaitGroup) ServiceID {
	svcID, newService := ParseService(k8sSvc, s.nodeAddressing)
	if newService == nil {
		return svcID
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldService, ok := s.services[svcID]
	if ok {
		if oldService.DeepEquals(newService) {
			return svcID
		}
	}
	s.services[svcID] = newService

	// Check if the corresponding Endpoints resource is already available
	endpoints, serviceReady := s.correlateEndpoints(svcID)
	if serviceReady {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:     UpdateService,
			ID:         svcID,
			Service:    newService,
			OldService: oldService,
			Endpoints:  endpoints,
			SWG:        swg,
		}
	}

	return svcID
}

// correlateEndpoints builds a combined Endpoints of the local endpoints and
// all external endpoints if the service is marked as a global service. Also
// returns a boolean that indicates whether the service is ready to be plumbed,
// this is true if:
// IF If ta local endpoints resource is present. Regardless whether the
//
//	endpoints resource contains actual backends or not.
//
// OR Remote endpoints exist which correlate to the service.
func (s *ServiceCache) correlateEndpoints(id ServiceID) (*Endpoints, bool) {
	endpoints := newEndpoints()
	localEndpoints := s.endpoints[id].GetEndpoints()
	hasLocalEndpoints := localEndpoints != nil
	if hasLocalEndpoints {
		for ip, e := range localEndpoints.Backends {
			endpoints.Backends[ip] = e
		}
	}

	svc, hasExternalService := s.services[id]
	if hasExternalService && svc.IncludeExternal {
		externalEndpoints, hasExternalEndpoints := s.externalEndpoints[id]
		if hasExternalEndpoints {
			// remote cluster endpoints already contain all Endpoints from all
			// EndpointSlices so no need to search the endpoints of a particular
			// EndpointSlice.
			for clusterName, remoteClusterEndpoints := range externalEndpoints.endpoints {
				if clusterName == option.Config.ClusterName {
					continue
				}

				for ip, e := range remoteClusterEndpoints.Backends {
					if _, ok := endpoints.Backends[ip]; ok {
						log.WithFields(log.Fields{
							logfields.K8sSvcName:   id.Name,
							logfields.K8sNamespace: id.Namespace,
							logfields.IPAddr:       ip,
							"cluster":              clusterName,
						}).Warning("Conflicting service backend IP")
					} else {
						endpoints.Backends[ip] = e
					}
				}
			}
		}
	}

	// Report the service as ready if a local endpoints object exists or if
	// external endpoints have have been identified
	return endpoints, hasLocalEndpoints || len(endpoints.Backends) > 0
}

func (s *ServiceCache) DeleteService(k8sSvc *corev1.Service, swg *lock.StoppableWaitGroup) {
	svcID := ParseServiceID(k8sSvc)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldService, serviceOK := s.services[svcID]
	endpoints, _ := s.correlateEndpoints(svcID)
	delete(s.services, svcID)

	if serviceOK {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:    DeleteService,
			ID:        svcID,
			Service:   oldService,
			Endpoints: endpoints,
			SWG:       swg,
		}
	}
}

func (s *ServiceCache) UpdateEndpoints(k8sEndpoints *corev1.Endpoints, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	svcID, newEndpoints := ParseEndpoints(k8sEndpoints)
	epSliceID := EndpointSliceID{
		ServiceID:         svcID,
		EndpointSliceName: k8sEndpoints.GetName(),
	}

	return s.updateEndpoints(epSliceID, newEndpoints, swg)
}

func (s *ServiceCache) UpdateEndpointSlices(epSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	svcID, newEndpoints := ParseEndpointSlice(epSlice)

	return s.updateEndpoints(svcID, newEndpoints, swg)
}

func (s *ServiceCache) updateEndpoints(esID EndpointSliceID, newEndpoints *Endpoints, swg *lock.StoppableWaitGroup) (ServiceID, *Endpoints) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	eps, ok := s.endpoints[esID.ServiceID]
	if ok {
		if eps.epSlices[esID.EndpointSliceName].DeepEquals(newEndpoints) {
			return esID.ServiceID, newEndpoints
		}
	} else {
		eps = newEndpointsSlices()
		s.endpoints[esID.ServiceID] = eps
	}

	eps.UpdateOrInsert(esID.EndpointSliceName, newEndpoints)

	// Check if the corresponding Endpoints resource is already available
	svc, ok := s.services[esID.ServiceID]
	endpoints, serviceReady := s.correlateEndpoints(esID.ServiceID)
	if ok && serviceReady {
		swg.Add()
		s.Events <- ServiceEvent{
			Action:    UpdateService,
			ID:        esID.ServiceID,
			Service:   svc,
			Endpoints: endpoints,
			SWG:       swg,
		}
	}

	return esID.ServiceID, newEndpoints
}

func (s *ServiceCache) DeleteEndpoints(k8sEndpoints *corev1.Endpoints, swg *lock.StoppableWaitGroup) ServiceID {
	svcID := ParseEndpointsID(k8sEndpoints)
	epSliceID := EndpointSliceID{
		ServiceID:         svcID,
		EndpointSliceName: k8sEndpoints.GetName(),
	}
	return s.deleteEndpoints(epSliceID, swg)
}

func (s *ServiceCache) deleteEndpoints(svcID EndpointSliceID, swg *lock.StoppableWaitGroup) ServiceID {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	svc, serviceOK := s.services[svcID.ServiceID]
	isEmpty := s.endpoints[svcID.ServiceID].Delete(svcID.EndpointSliceName)
	if isEmpty {
		delete(s.endpoints, svcID.ServiceID)
	}
	endpoints, _ := s.correlateEndpoints(svcID.ServiceID)

	if serviceOK {
		swg.Add()
		event := ServiceEvent{
			Action:    UpdateService,
			ID:        svcID.ServiceID,
			Service:   svc,
			Endpoints: endpoints,
			SWG:       swg,
		}

		s.Events <- event
	}

	return svcID.ServiceID
}

func (s *ServiceCache) DeleteEndpointSlices(epSlice *discoveryv1.EndpointSlice, swg *lock.StoppableWaitGroup) ServiceID {
	svcID := ParseEndpointSliceID(epSlice)

	return s.deleteEndpoints(svcID, swg)
}
