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
