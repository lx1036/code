package storage

import (
	"sync"

	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

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

	for _, nodePoint := range batch.Nodes {
		if _, exists := newNodes[nodePoint.Name]; exists {
			klog.Errorf("duplicate node %s received", nodePoint.Name)
			continue
		}
		newNodes[nodePoint.Name] = nodePoint
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
