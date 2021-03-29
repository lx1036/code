package storage

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/monitor/metrcis-server/pkg/api"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics"
)

// kubernetesCadvisorWindow is the max window used by cAdvisor for calculating
// CPU usage rate.  While it can vary, it's no more than this number, but may be
// as low as half this number (when working with no backoff).  It would be really
// nice if the kubelet told us this in the summary API...
var kubernetesCadvisorWindow = 30 * time.Second

type Storage struct {
	mu    sync.RWMutex
	nodes map[string]NodeMetricsPoint                 // 哈希表存储 node metrics 数据
	pods  map[apitypes.NamespacedName]PodMetricsPoint // 哈希表存储 pod metrics 数据
}

func NewStorage() *Storage {
	return &Storage{
		nodes: make(map[string]NodeMetricsPoint),
		pods:  make(map[apitypes.NamespacedName]PodMetricsPoint),
	}
}

func (storage *Storage) Store(batch *MetricsBatch) {
	newNodes := make(map[string]NodeMetricsPoint, len(batch.Nodes))
	newPods := make(map[apitypes.NamespacedName]PodMetricsPoint, len(batch.Pods))

	for _, node := range batch.Nodes {
		if _, exists := newNodes[node.Name]; exists {
			klog.Errorf("duplicate node %s received", node.Name)
			continue
		}
		newNodes[node.Name] = node
	}

	for _, podPoint := range batch.Pods {
		podIdent := apitypes.NamespacedName{Name: podPoint.Name, Namespace: podPoint.Namespace}
		if _, exists := newPods[podIdent]; exists {
			klog.Errorf("duplicate pod %s received", podIdent)
			continue
		}
		newPods[podIdent] = podPoint
	}

	storage.mu.Lock()
	storage.nodes = newNodes
	storage.pods = newPods
	storage.mu.Unlock()
}

func (storage *Storage) GetContainerMetrics(pods ...apitypes.NamespacedName) ([]api.TimeInfo, [][]metrics.ContainerMetrics) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	timestamps := make([]api.TimeInfo, len(pods))
	resMetrics := make([][]metrics.ContainerMetrics, len(pods))

	for i, pod := range pods {
		metricPoint, present := storage.pods[pod]
		if !present {
			continue
		}

		contMetrics := make([]metrics.ContainerMetrics, len(metricPoint.Containers))
		var earliestTS *time.Time
		for key, contPoint := range metricPoint.Containers {
			contMetrics[key] = metrics.ContainerMetrics{
				Name: contPoint.Name,
				Usage: corev1.ResourceList{
					corev1.ResourceCPU:    contPoint.CpuUsage,
					corev1.ResourceMemory: contPoint.MemoryUsage,
				},
			}
			if earliestTS == nil || earliestTS.After(contPoint.Timestamp) {
				ts := contPoint.Timestamp // copy to avoid loop iteration variable issues
				earliestTS = &ts
			}
		}
		if earliestTS == nil {
			// we had no containers
			earliestTS = &time.Time{}
		}
		timestamps[i] = api.TimeInfo{
			Timestamp: *earliestTS,
			Window:    kubernetesCadvisorWindow,
		}
		resMetrics[i] = contMetrics
	}

	return timestamps, resMetrics
}

func (storage *Storage) GetNodeMetrics(nodes ...string) ([]api.TimeInfo, []corev1.ResourceList) {
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	timestamps := make([]api.TimeInfo, len(nodes))
	resMetrics := make([]corev1.ResourceList, len(nodes))

	for i, node := range nodes {
		metricPoint, present := storage.nodes[node]
		if !present {
			continue
		}

		timestamps[i] = api.TimeInfo{
			Timestamp: metricPoint.Timestamp,
			Window:    kubernetesCadvisorWindow,
		}
		resMetrics[i] = corev1.ResourceList{
			corev1.ResourceCPU:    metricPoint.CpuUsage,
			corev1.ResourceMemory: metricPoint.MemoryUsage,
		}
	}

	return timestamps, resMetrics
}
