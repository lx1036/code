package pod

import (
	"golang.org/x/build/kubernetes/api"
	"k8s-lx1036/k8s-ui/backend/common/kubeclient"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func GetPodListByType(kubeClient kubeclient.ResourceHandler, namespace, resourceName string, resourceType api.ResourceName) ([]*v1.Pod, error) {
	switch resourceType {
	case api.ResourcePods:
		obj, err := kubeClient.Get(string(api.ResourcePods), namespace, resourceName)
		if err != nil {
			return nil, err
		}
		relatePod := []*v1.Pod{
			obj.(*v1.Pod),
		}
		return relatePod, nil
	default:
		return nil, nil
	}
}

// GetPodStatus returns the pod state
func GetPodStatus(pod *v1.Pod) string {
	// Terminating
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	// not running
	if pod.Status.Phase != v1.PodRunning {
		return string(pod.Status.Phase)
	}

	ready := false
	notReadyReason := ""
	for _, c := range pod.Status.Conditions {
		if c.Type == v1.PodReady {
			ready = c.Status == v1.ConditionTrue
			notReadyReason = c.Reason
		}
	}

	if pod.Status.Reason != "" {
		return pod.Status.Reason
	}

	if notReadyReason != "" {
		return notReadyReason
	}

	if ready {
		return string(v1.PodRunning)
	}

	// Unknown?
	return "Unknown"
}

func GetPodCounts(cache *kubeclient.CacheFactory) (int, error) {
	pods, err := cache.PodLister().List(labels.Everything())
	if err != nil {
		return 0, nil
	}
	length := 0
	for _, pod := range pods {
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		length++
	}
	return length, nil
}
