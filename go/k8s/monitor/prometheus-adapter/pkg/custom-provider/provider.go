package custom_provider

import (
	"github.com/kubernetes-sigs/custom-metrics-apiserver/pkg/provider"
	"k8s.io/client-go/dynamic"
)

func NewPrometheusProvider(mapper apimeta.RESTMapper,
	kubeClient dynamic.Interface,
	promClient prom.Client,
	namers []naming.MetricNamer,
	updateInterval time.Duration,
	maxAge time.Duration) (provider.CustomMetricsProvider, Runnable) {

}
