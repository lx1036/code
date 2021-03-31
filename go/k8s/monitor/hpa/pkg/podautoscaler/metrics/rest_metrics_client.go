package metrics

import (
	"context"
	"fmt"
	"k8s.io/klog/v2"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
	customclient "k8s.io/metrics/pkg/client/custom_metrics"
	externalclient "k8s.io/metrics/pkg/client/external_metrics"
)

// PodMetric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type PodMetric struct {
	Timestamp time.Time
	Window    time.Duration
	Value     int64
}

// PodMetricsInfo contains pod metrics as a map from pod names to PodMetricsInfo
type PodMetricsInfo map[string]PodMetric

// restMetricsClient is a client which supports fetching
// metrics from both the resource metrics API and the
// custom metrics API.
type RestMetricsClient struct {
	*resourceMetricsClient
	*customMetricsClient
	*externalMetricsClient
}

type resourceMetricsClient struct {
	client resourceclient.PodMetricsesGetter
}

type customMetricsClient struct {
	client customclient.CustomMetricsClient
}

type externalMetricsClient struct {
	client externalclient.ExternalMetricsClient
}

func NewRESTMetricsClient(resourceClient resourceclient.PodMetricsesGetter,
	customClient customclient.CustomMetricsClient,
	externalClient externalclient.ExternalMetricsClient) *RestMetricsClient {
	return &RestMetricsClient{
		&resourceMetricsClient{resourceClient},
		&customMetricsClient{customClient},
		&externalMetricsClient{externalClient},
	}
}

func (c *resourceMetricsClient) GetResourceMetric(resource v1.ResourceName,
	namespace string, selector labels.Selector) (PodMetricsInfo, time.Time, error) {
	metrics, err := c.client.PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("unable to fetch metrics from resource metrics API: %v", err)
	}
	if len(metrics.Items) == 0 {
		return nil, time.Time{}, fmt.Errorf("no metrics returned from resource metrics API")
	}

	res := make(PodMetricsInfo, len(metrics.Items))
	for _, m := range metrics.Items {
		podSum := int64(0)
		missing := len(m.Containers) == 0
		for _, c := range m.Containers {
			resValue, found := c.Usage[resource]
			if !found {
				missing = true
				klog.Infof("missing resource metric %v for container %s in pod %s/%s", resource, c.Name, namespace, m.Name)
				break // containers loop
			}
			podSum += resValue.MilliValue()
		}

		if !missing {
			res[m.Name] = PodMetric{
				Timestamp: m.Timestamp.Time,
				Window:    m.Window.Duration,
				Value:     int64(podSum),
			}
		}
	}

	timestamp := metrics.Items[0].Timestamp.Time
	return res, timestamp, nil
}
