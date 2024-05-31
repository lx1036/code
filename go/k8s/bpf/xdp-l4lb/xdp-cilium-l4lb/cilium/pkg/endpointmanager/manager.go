package endpointmanager

import (
    "fmt"
    "k8s-lx1036/k8s/network/cilium/cilium/pkg/metrics"
    "sync"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint"
    endpointid "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/id"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"

    "github.com/prometheus/client_golang/prometheus"
)

var (
    metricsOnce sync.Once
)

// EndpointManager is a structure designed for containing state about the
// collection of locally running endpoints.
type EndpointManager struct {
    mutex sync.RWMutex

    // endpoints is the global list of endpoints indexed by ID. mutex must
    // be held to read and write.
    endpoints    map[uint16]*endpoint.Endpoint
    endpointsAux map[string]*endpoint.Endpoint
}

func NewEndpointManager(epSynchronizer EndpointResourceSynchronizer) *EndpointManager {
    mgr := EndpointManager{
        endpoints:                    make(map[uint16]*endpoint.Endpoint),
        endpointsAux:                 make(map[string]*endpoint.Endpoint),
        EndpointResourceSynchronizer: epSynchronizer,
    }

    return &mgr
}

func (mgr *EndpointManager) InitMetrics() {
    metricsOnce.Do(func() { // EndpointCount is a function used to collect this metric. We cannot
        // increment/decrement a gauge since we invoke Remove gratuitiously and that
        // would result in negative counts.
        // It must be thread-safe.
        metrics.EndpointCount = prometheus.NewGaugeFunc(prometheus.GaugeOpts{
            Namespace: metrics.Namespace,
            Name:      "endpoint_count",
            Help:      "Number of endpoints managed by this agent",
        },
            func() float64 { return float64(len(mgr.GetEndpoints())) },
        )

        metrics.MustRegister(metrics.EndpointCount)
    })
}

// GetHostEndpoint returns the host endpoint.
func (mgr *EndpointManager) GetHostEndpoint() *endpoint.Endpoint {
    mgr.mutex.RLock()
    defer mgr.mutex.RUnlock()
    for _, ep := range mgr.endpoints {
        if ep.IsHost() {
            return ep
        }
    }
    return nil
}

// GetEndpoints returns a slice of all endpoints present in endpoint manager.
func (mgr *EndpointManager) GetEndpoints() []*endpoint.Endpoint {
    mgr.mutex.RLock()
    eps := make([]*endpoint.Endpoint, 0, len(mgr.endpoints))
    for _, ep := range mgr.endpoints {
        eps = append(eps, ep)
    }
    mgr.mutex.RUnlock()
    return eps
}

// Lookup looks up the endpoint by prefix id
func (mgr *EndpointManager) Lookup(id string) (*endpoint.Endpoint, error) {
    mgr.mutex.RLock()
    defer mgr.mutex.RUnlock()

    prefix, eid, err := endpointid.Parse(id)
    if err != nil {
        return nil, err
    }

    switch prefix {
    case endpointid.CiliumLocalIdPrefix:
        n, err := endpointid.ParseCiliumID(id)
        if err != nil {
            return nil, err
        }
        if n > endpointid.MaxEndpointId {
            return nil, fmt.Errorf("%d: endpoint ID too large", n)
        }
        return mgr.lookupCiliumID(uint16(n)), nil

    case endpointid.CiliumGlobalIdPrefix:
        return nil, ErrUnsupportedID

    case endpointid.ContainerIdPrefix:
        return mgr.lookupContainerID(eid), nil

    case endpointid.DockerEndpointPrefix:
        return mgr.lookupDockerEndpoint(eid), nil

    case endpointid.ContainerNamePrefix:
        return mgr.lookupDockerContainerName(eid), nil

    case endpointid.PodNamePrefix:
        return mgr.lookupPodNameLocked(eid), nil

    case endpointid.IPv4Prefix:
        return mgr.lookupIPv4(eid), nil

    //case endpointid.IPv6Prefix:
    //    return mgr.lookupIPv6(eid), nil

    default:
        return nil, ErrInvalidPrefix{InvalidPrefix: prefix.String()}
    }
}

func (mgr *EndpointManager) lookupCiliumID(id uint16) *endpoint.Endpoint {
    if ep, ok := mgr.endpoints[id]; ok {
        return ep
    }

    return nil
}

func (mgr *EndpointManager) lookupDockerEndpoint(id string) *endpoint.Endpoint {
    if ep, ok := mgr.endpointsAux[endpointid.NewID(endpointid.DockerEndpointPrefix, id)]; ok {
        return ep
    }
    return nil
}

func (mgr *EndpointManager) lookupPodNameLocked(name string) *endpoint.Endpoint {
    if ep, ok := mgr.endpointsAux[endpointid.NewID(endpointid.PodNamePrefix, name)]; ok {
        return ep
    }
    return nil
}

func (mgr *EndpointManager) lookupDockerContainerName(name string) *endpoint.Endpoint {
    if ep, ok := mgr.endpointsAux[endpointid.NewID(endpointid.ContainerNamePrefix, name)]; ok {
        return ep
    }
    return nil
}

func (mgr *EndpointManager) lookupIPv4(ipv4 string) *endpoint.Endpoint {
    if ep, ok := mgr.endpointsAux[endpointid.NewID(endpointid.IPv4Prefix, ipv4)]; ok {
        return ep
    }
    return nil
}

func (mgr *EndpointManager) lookupContainerID(id string) *endpoint.Endpoint {
    if ep, ok := mgr.endpointsAux[endpointid.NewID(endpointid.ContainerIdPrefix, id)]; ok {
        return ep
    }
    return nil
}

// AddEndpoint takes the prepared endpoint object and starts managing it.
func (mgr *EndpointManager) AddEndpoint(owner regeneration.Owner, ep *endpoint.Endpoint, reason string) (err error) {

}
