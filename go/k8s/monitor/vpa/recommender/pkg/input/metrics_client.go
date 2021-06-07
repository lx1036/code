package input

import (
	"context"
	"time"

	"k8s-lx1036/k8s/monitor/vpa/recommender/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	resourceclient "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type MetricsClient struct {
	metricsGetter resourceclient.PodMetricsesGetter
	namespace     string
}

// ContainerMetricsSnapshot contains information about usage of certain container within defined time window.
type ContainerMetricsSnapshot struct {
	// ID identifies a specific container those metrics are coming from.
	ID types.ContainerID
	// End time of the measurement interval.
	SnapshotTime time.Time
	// Duration of the measurement interval, which is [SnapshotTime - SnapshotWindow, SnapshotTime].
	SnapshotWindow time.Duration
	// Actual usage of the resources over the measurement interval.
	Usage types.Resources
}

func NewMetricsClient(metricsGetter resourceclient.PodMetricsesGetter, namespace string) *MetricsClient {
	return &MetricsClient{
		metricsGetter: metricsGetter,
		namespace:     namespace,
	}
}

func (c *MetricsClient) GetContainersMetrics() ([]*ContainerMetricsSnapshot, error) {
	var metricsSnapshots []*ContainerMetricsSnapshot

	podMetricsInterface := c.metricsGetter.PodMetricses(c.namespace)
	podMetricsList, err := podMetricsInterface.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	klog.V(3).Infof("%v podMetrics retrieved for all namespaces", len(podMetricsList.Items))
	for _, podMetrics := range podMetricsList.Items {
		metricsSnapshotsForPod := createContainerMetricsSnapshots(podMetrics)
		metricsSnapshots = append(metricsSnapshots, metricsSnapshotsForPod...)
	}

	return metricsSnapshots, nil
}

func createContainerMetricsSnapshots(podMetrics v1beta1.PodMetrics) []*ContainerMetricsSnapshot {
	snapshots := make([]*ContainerMetricsSnapshot, len(podMetrics.Containers))
	for i, containerMetrics := range podMetrics.Containers {
		snapshots[i] = newContainerMetricsSnapshot(containerMetrics, podMetrics)
	}
	return snapshots
}

func newContainerMetricsSnapshot(containerMetrics v1beta1.ContainerMetrics, podMetrics v1beta1.PodMetrics) *ContainerMetricsSnapshot {
	usage := calculateUsage(containerMetrics.Usage)

	return &ContainerMetricsSnapshot{
		ID: types.ContainerID{
			ContainerName: containerMetrics.Name,
			PodID: types.PodID{
				Namespace: podMetrics.Namespace,
				PodName:   podMetrics.Name,
			},
		},
		Usage:          usage,
		SnapshotTime:   podMetrics.Timestamp.Time,
		SnapshotWindow: podMetrics.Window.Duration,
	}
}

func calculateUsage(containerUsage corev1.ResourceList) types.Resources {
	cpuQuantity := containerUsage[corev1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue() // cpu: e.g. 1234m
	memoryQuantity := containerUsage[corev1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return types.Resources{
		types.ResourceCPU:    types.ResourceAmount(cpuMillicores),
		types.ResourceMemory: types.ResourceAmount(memoryBytes),
	}
}
