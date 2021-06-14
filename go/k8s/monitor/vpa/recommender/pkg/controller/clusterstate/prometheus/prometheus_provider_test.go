package prometheus

import (
	"k8s.io/klog/v2"
	"testing"
	"time"
)

func TestPrometheusHistoryProvider(test *testing.T) {
	config := PrometheusHistoryProviderConfig{
		Address:           "http://localhost:9090",
		QueryTimeout:      time.Minute * 5,
		HistoryLength:     "1d",
		HistoryResolution: "1h",

		PodLabelPrefix:      "pod_label_",
		PodLabelsMetricName: `up{job="kubernetes-custom"}`,
		PodNamespaceLabel:   "kubernetes_namespace",
		PodNameLabel:        "pod",

		CadvisorMetricsJobName: "cadvisor",

		ContainerNameLabel:      "name",
		ContainerPodNameLabel:   "pod",
		ContainerNamespaceLabel: "namespace",

		//Namespace: corev1.NamespaceAll,
		Namespace: "cattle-system",
	}
	provider, err := NewPrometheusHistoryProvider(config)
	if err != nil {
		klog.Fatalf("Could not initialize history provider: %v", err)
	}

	clusterHistory, err := provider.GetClusterHistory()
	if err != nil {
		klog.Fatalf("Cannot get cluster history: %v", err)
	}

	for podID, podHistory := range clusterHistory {
		klog.Info(podID, podHistory.Samples)
	}

}
