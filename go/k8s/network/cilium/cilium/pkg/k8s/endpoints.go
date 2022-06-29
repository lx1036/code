package k8s

import (
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/k8s/loadbalancer"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

// EndpointSliceID identifies a Kubernetes EndpointSlice as well as the legacy
// v1.Endpoints.
type EndpointSliceID struct {
	ServiceID
	EndpointSliceName string
}

// endpointSlices is the collection of all endpoint slices of a service.
// The map key is the name of the endpoint slice or the name of the legacy
// v1.Endpoint. The endpoints stored here are not namespaced since this
// structure is only used as a value of another map that is already namespaced.
// (see ServiceCache.endpoints).
type endpointSlices struct {
	epSlices map[string]*Endpoints
}

// newEndpointsSlices returns a new endpointSlices
func newEndpointsSlices() *endpointSlices {
	return &endpointSlices{
		epSlices: map[string]*Endpoints{},
	}
}

func (es *endpointSlices) UpdateOrInsert(esName string, e *Endpoints) {
	if es == nil {
		panic("BUG: endpointSlices is nil")
	}
	es.epSlices[esName] = e
}

func (es *endpointSlices) GetEndpoints() *Endpoints {
	if es == nil || len(es.epSlices) == 0 {
		return nil
	}
	allEps := newEndpoints()
	for _, eps := range es.epSlices {
		for backend, ep := range eps.Backends {
			allEps.Backends[backend] = ep
		}
	}
	return allEps
}

// Endpoints is an abstraction for the Kubernetes endpoints object. Endpoints
// consists of a set of backend IPs in combination with a set of ports and
// protocols. The name of the backend ports must match the names of the
// frontend ports of the corresponding service.
// +k8s:deepcopy-gen=true
type Endpoints struct {
	// Backends is a map containing all backend IPs and ports. The key to
	// the map is the backend IP in string form. The value defines the list
	// of ports for that backend IP, plus an additional optional node name.
	Backends map[string]*Backend
}

func newEndpoints() *Endpoints {
	return &Endpoints{
		Backends: map[string]*Backend{},
	}
}

// Backend contains all ports and the node name of a given backend
// +k8s:deepcopy-gen=true
type Backend struct {
	Ports    PortConfiguration
	NodeName string
}

// externalEndpoints is the collection of external endpoints in all remote
// clusters. The map key is the name of the remote cluster.
type externalEndpoints struct {
	endpoints map[string]*Endpoints
}

// ParseEndpoints parses a Kubernetes Endpoints resource
func ParseEndpoints(ep *corev1.Endpoints) (ServiceID, *Endpoints) {
	endpoints := newEndpoints()

	for _, sub := range ep.Subsets {
		for _, addr := range sub.Addresses {
			backend, ok := endpoints.Backends[addr.IP]
			if !ok {
				backend = &Backend{Ports: PortConfiguration{}}
				endpoints.Backends[addr.IP] = backend
			}

			if addr.NodeName != nil {
				backend.NodeName = *addr.NodeName
			}

			for _, port := range sub.Ports {
				lbPort := loadbalancer.NewL4Addr(loadbalancer.L4Type(port.Protocol), uint16(port.Port))
				backend.Ports[port.Name] = lbPort
			}
		}
	}

	return ParseEndpointsID(ep), endpoints
}

func ParseEndpointsID(svc *corev1.Endpoints) ServiceID {
	return ServiceID{
		Name:      svc.ObjectMeta.Name,
		Namespace: svc.ObjectMeta.Namespace,
	}
}

// ParseEndpointSlice parses a Kubernetes Endpoints resource
func ParseEndpointSlice(ep *discoveryv1.EndpointSlice) (EndpointSliceID, *Endpoints) {
	endpoints := newEndpoints()

	for _, sub := range ep.Endpoints {
		// ready indicates that this endpoint is prepared to receive traffic,
		// according to whatever system is managing the endpoint. A nil value
		// indicates an unknown state. In most cases consumers should interpret this
		// unknown state as ready.
		// More info: vendor/k8s.io/api/discovery/v1beta1/types.go:114
		if sub.Conditions.Ready != nil && !*sub.Conditions.Ready {
			continue
		}
		for _, addr := range sub.Addresses {
			backend, ok := endpoints.Backends[addr]
			if !ok {
				backend = &Backend{Ports: PortConfiguration{}}
				endpoints.Backends[addr] = backend

				if sub.NodeName != nil {
					backend.NodeName = *sub.NodeName
				}
			}

			for _, port := range ep.Ports {
				name, lbPort := parseEndpointPort(port)
				if lbPort != nil {
					backend.Ports[name] = lbPort
				}
			}
		}
	}

	return ParseEndpointSliceID(ep), endpoints
}

func parseEndpointPort(port discoveryv1.EndpointPort) (string, *loadbalancer.L4Addr) {
	proto := loadbalancer.TCP
	if port.Protocol != nil {
		switch *port.Protocol {
		case corev1.ProtocolTCP:
			proto = loadbalancer.TCP
		case corev1.ProtocolUDP:
			proto = loadbalancer.UDP
		default:
			return "", nil
		}
	}
	if port.Port == nil {
		return "", nil
	}
	var name string
	if port.Name != nil {
		name = *port.Name
	}
	lbPort := loadbalancer.NewL4Addr(proto, uint16(*port.Port))
	return name, lbPort
}

func ParseEndpointSliceID(es *discoveryv1.EndpointSlice) EndpointSliceID {
	return EndpointSliceID{
		ServiceID: ServiceID{
			Name:      es.ObjectMeta.GetLabels()[discoveryv1.LabelServiceName],
			Namespace: es.ObjectMeta.Namespace,
		},
		EndpointSliceName: es.ObjectMeta.GetName(),
	}
}
