package podautoscaler

import (
	"fmt"
	"k8s-lx1036/k8s/monitor/hpa/pkg/podautoscaler/metrics"
	v1 "k8s.io/api/core/v1"
)

func calculatePodRequests(pods []*v1.Pod, resource v1.ResourceName) (map[string]int64, error) {
	requests := make(map[string]int64, len(pods))
	for _, pod := range pods {
		podSum := int64(0)
		for _, container := range pod.Spec.Containers {
			if containerRequest, ok := container.Resources.Requests[resource]; ok {
				podSum += containerRequest.MilliValue()
			} else {
				return nil, fmt.Errorf("missing request for %s", resource)
			}
		}
		requests[pod.Name] = podSum
	}
	
	return requests, nil
}



func removeMetricsForPods(metrics metrics.PodMetricsInfo, pods sets.String) {
	for _, pod := range pods.UnsortedList() {
		delete(metrics, pod)
	}
}

