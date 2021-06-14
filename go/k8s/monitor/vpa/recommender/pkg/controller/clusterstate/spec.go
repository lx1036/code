package clusterstate

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

// BasicPodSpec contains basic information defining a pod and its containers.
type BasicPodSpec struct {
	// ID identifies a pod within a cluster.
	ID PodID
	// Labels of the pod. It is used to match pods with certain VPA opjects.
	PodLabels map[string]string
	// List of containers within this pod.
	Containers []BasicContainerSpec
	// PodPhase describing current life cycle phase of the Pod.
	Phase v1.PodPhase
}

// BasicContainerSpec contains basic information defining a container.
type BasicContainerSpec struct {
	// ID identifies the container within a cluster.
	ID ContainerID
	// Name of the image running within the container.
	Image string
	// Currently requested resources for this container.
	Request Resources
}

type SpecClient struct {
	podLister listersv1.PodLister
}

// NewSpecClient creates new client which can be used to get basic information about pods specification
// It requires PodLister which is a data source for this client.
func NewSpecClient(podLister listersv1.PodLister) *SpecClient {
	return &SpecClient{
		podLister: podLister,
	}
}

func (client *SpecClient) GetPodSpecs() ([]*BasicPodSpec, error) {
	var podSpecs []*BasicPodSpec

	pods, err := client.podLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, pod := range pods {
		basicPodSpec := newBasicPodSpec(pod)
		podSpecs = append(podSpecs, basicPodSpec)
	}
	return podSpecs, nil
}

func newBasicPodSpec(pod *v1.Pod) *BasicPodSpec {
	podId := PodID{
		PodName:   pod.Name,
		Namespace: pod.Namespace,
	}
	containerSpecs := newContainerSpecs(podId, pod)

	basicPodSpec := &BasicPodSpec{
		ID:         podId,
		PodLabels:  pod.Labels,
		Containers: containerSpecs,
		Phase:      pod.Status.Phase,
	}

	return basicPodSpec
}

func newContainerSpecs(podID PodID, pod *v1.Pod) []BasicContainerSpec {
	var containerSpecs []BasicContainerSpec

	for _, container := range pod.Spec.Containers {
		containerSpec := newContainerSpec(podID, container)
		containerSpecs = append(containerSpecs, containerSpec)
	}

	return containerSpecs
}

func newContainerSpec(podID PodID, container v1.Container) BasicContainerSpec {
	containerSpec := BasicContainerSpec{
		ID: ContainerID{
			PodID:         podID,
			ContainerName: container.Name,
		},
		Image:   container.Image,
		Request: calculateRequestedResources(container),
	}

	return containerSpec
}

func calculateRequestedResources(container v1.Container) Resources {
	cpuQuantity := container.Resources.Requests[v1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := container.Resources.Requests[v1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return Resources{
		ResourceCPU:    ResourceAmount(cpuMillicores),
		ResourceMemory: ResourceAmount(memoryBytes),
	}

}
