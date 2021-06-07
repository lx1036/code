package input

import resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"


type MetricsClient struct {
	metricsGetter resourceclient.PodMetricsesGetter
	namespace     string
}


func NewMetricsClient(metricsGetter resourceclient.PodMetricsesGetter, namespace string) *MetricsClient {
	return &MetricsClient{
		metricsGetter: metricsGetter,
		namespace:     namespace,
	}
}
