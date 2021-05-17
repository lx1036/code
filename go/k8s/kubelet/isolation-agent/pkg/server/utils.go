package server

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
)

// INFO: @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/kubelet_pods.go#L96-L101
func GetActivePods(allPods []*v1.Pod) []*v1.Pod {
	activePods := filterOutTerminatedPods(allPods)
	return activePods
}

// filterOutTerminatedPods returns the given pods which the status manager
// does not consider failed or succeeded.
func filterOutTerminatedPods(pods []*v1.Pod) []*v1.Pod {
	var filteredPods []*v1.Pod
	for _, p := range pods {
		if podIsTerminated(p) {
			continue
		}
		filteredPods = append(filteredPods, p)
	}
	return filteredPods
}

// podIsTerminated returns true if the provided pod is in a terminal phase ("Failed", "Succeeded") or
// has been deleted and has no running containers. This corresponds to when a pod must accept changes to
// its pod spec (e.g. terminating containers allow grace period to be shortened).
func podIsTerminated(pod *v1.Pod) bool {
	_, podWorkerTerminal := podAndContainersAreTerminal(pod)
	return podWorkerTerminal
}

// podStatusIsTerminal reports when the specified pod has no running containers or is no longer accepting
// spec changes.
func podAndContainersAreTerminal(pod *v1.Pod) (containersTerminal, podWorkerTerminal bool) {
	status := pod.Status

	// A pod transitions into failed or succeeded from either container lifecycle (RestartNever container
	// fails) or due to external events like deletion or eviction. A terminal pod *should* have no running
	// containers, but to know that the pod has completed its lifecycle you must wait for containers to also
	// be terminal.
	containersTerminal = notRunning(status.ContainerStatuses)
	// The kubelet must accept config changes from the pod spec until it has reached a point where changes would
	// have no effect on any running container.
	podWorkerTerminal = status.Phase == v1.PodFailed || status.Phase == v1.PodSucceeded || (pod.DeletionTimestamp != nil && containersTerminal)
	return
}

// notRunning returns true if every status is terminated or waiting, or the status list
// is empty.
func notRunning(statuses []v1.ContainerStatus) bool {
	for _, status := range statuses {
		if status.State.Terminated == nil && status.State.Waiting == nil {
			return false
		}
	}
	return true
}

// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/cm/cpumanager/cpu_manager.go#L431-L445
func findContainerIDByName(status *corev1.PodStatus, name string) (string, error) {
	allStatuses := status.InitContainerStatuses
	allStatuses = append(allStatuses, status.ContainerStatuses...)
	for _, container := range allStatuses {
		if container.Name == name && container.ContainerID != "" {
			cid := &kubecontainer.ContainerID{}
			err := cid.ParseString(container.ContainerID)
			if err != nil {
				return "", err
			}
			return cid.ID, nil
		}
	}

	return "", fmt.Errorf("unable to find ID for container with name %v in pod status (it may not be running)", name)
}

// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/cm/cpumanager/cpu_manager.go#L447-L454
func findContainerStatusByName(status *corev1.PodStatus, name string) (*corev1.ContainerStatus, error) {
	for _, status := range append(status.InitContainerStatuses, status.ContainerStatuses...) {
		if status.Name == name {
			return &status, nil
		}
	}

	return nil, fmt.Errorf("unable to find status for container with name %v in pod status (it may not be running)", name)
}

func NewRestConfig(kubeconfig string) (*rest.Config, error) {
	var config *rest.Config
	if _, err := os.Stat(kubeconfig); err == nil {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
