package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	metricsControllerTickTime = 3 * time.Second
	namespace                 = "kube_router"
)

var (
	// ControllerBGPadvertisementsReceived Time it took to sync internal bgp peers
	ControllerBGPAdvertisementsReceived = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "controller_bgp_advertisements_received",
		Help:      "BGP advertisements received",
	})

	// ControllerBGPAdvertisementsSent Time it took to sync internal bgp peers
	ControllerBGPAdvertisementsSent = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "controller_bgp_advertisements_sent",
		Help:      "BGP advertisements sent",
	})

	// ControllerBPGPeers BGP peers in the runtime configuration
	ControllerBPGPeers = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "controller_bgp_peers",
		Help:      "BGP peers in the runtime configuration",
	})
)
