package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	registry = prometheus.NewPedanticRegistry()
)

var (
	EndpointCount prometheus.GaugeFunc
)

// MustRegister It will panic on error.
func MustRegister(c ...prometheus.Collector) {
	registry.MustRegister(c...)
}
