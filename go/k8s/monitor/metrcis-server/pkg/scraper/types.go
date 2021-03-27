package scraper

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Summary struct {
	Node NodeStats `json:"node"`

	Pods []PodStats `json:"pods"`
}

type NodeStats struct {
	NodeName string `json:"nodeName"`

	CPU *CPUStats `json:"cpu,omitempty"`

	Memory *MemoryStats `json:"memory,omitempty"`
}

type PodStats struct {
	PodRef PodReference `json:"podRef"`

	Containers []ContainerStats `json:"containers" patchStrategy:"merge" patchMergeKey:"name"`
}

type ContainerStats struct {
	Name   string       `json:"name"`
	CPU    *CPUStats    `json:"cpu,omitempty"`
	Memory *MemoryStats `json:"memory,omitempty"`
}

type PodReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type CPUStats struct {
	Time           metav1.Time `json:"time"`
	UsageNanoCores *uint64     `json:"usageNanoCores,omitempty"`
}

// MemoryStats contains data about memory usage.
type MemoryStats struct {
	Time            metav1.Time `json:"time"`
	WorkingSetBytes *uint64     `json:"workingSetBytes,omitempty"`
}

/*
{
	"kind": "PodMetrics",
	"apiVersion": "metrics.k8s.io/v1beta1",
	"metadata": {
		"name": "exporter-node-cluster-monitoring-7zrjz",
		"namespace": "cattle-prometheus",
		"selfLink": "/apis/metrics.k8s.io/v1beta1/namespaces/cattle-prometheus/pods/exporter-node-cluster-monitoring-7zrjz",
		"creationTimestamp": "2021-03-27T15:22:31Z"
	},
	"timestamp": "2021-03-27T15:22:18Z",
	"window": "30s",
	"containers": [
		{
			"name": "exporter-node",
			"usage": {
				"cpu": "15235900n",
				"memory": "22868Ki"
			}
		}
	]
}
*/

/*
{
	"kind": "NodeMetrics",
	"apiVersion": "metrics.k8s.io/v1beta1",
	"metadata": {
		"name": "docker01.node",
		"selfLink": "/apis/metrics.k8s.io/v1beta1/nodes/docker01.node",
		"creationTimestamp": "2021-03-27T15:58:37Z"
	},
	"timestamp": "2021-03-27T15:58:23Z",
	"window": "30s",
	"usage": {
		"cpu": "232250128n",
		"memory": "4158324Ki"
	}
}
*/
