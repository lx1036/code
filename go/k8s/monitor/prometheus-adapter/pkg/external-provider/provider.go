package external_provider

import "github.com/kubernetes-sigs/custom-metrics-apiserver/pkg/provider"

// NewExternalPrometheusProvider creates an ExternalMetricsProvider capable of responding to Kubernetes requests for external metric data
func NewExternalPrometheusProvider(promClient prom.Client,
	namers []naming.MetricNamer,
	updateInterval time.Duration) (provider.ExternalMetricsProvider, Runnable) {

}
