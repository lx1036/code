package pod

import (
	"k8s-lx1036/k8s-ui/backend/client"
	"k8s-lx1036/k8s-ui/backend/client/api"
	v1 "k8s.io/api/core/v1"
)

func GetPodListByType(kubeClient client.ResourceHandler, namespace, resourceName string, resourceType api.ResourceName) ([]*v1.Pod, error) {
	switch resourceType {
	case api.ResourceNameDeployment:
	case api.ResourceNameCronJob:
	case api.ResourceNameDaemonSet, api.ResourceNameStatefulSet, api.ResourceNameJob:
	case api.ResourceNamePod:
		obj, err := kubeClient.Get(api.ResourceNamePod, namespace, resourceName)
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

	return nil, nil
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
