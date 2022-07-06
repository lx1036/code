package endpointmanager

import (
	"sync"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf/endpoint"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/metrics"

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
	endpoints map[uint16]*endpoint.Endpoint
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
