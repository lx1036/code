package devicemanager

import (
	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"
	"k8s-lx1036/k8s/kubelet/pkg/lifecycle"

	v1 "k8s.io/api/core/v1"
	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/config"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// Manager manages all the Device Plugins running on a node.
type Manager interface {
	// Start starts device plugin registration service.
	Start(activePods ActivePodsFunc, sourcesReady config.SourcesReady) error

	// Allocate configures and assigns devices to a container in a pod. From
	// the requested device resources, Allocate will communicate with the
	// owning device plugin to allow setup procedures to take place, and for
	// the device plugin to provide runtime settings to use the device
	// (environment variables, mount points and device files).
	Allocate(pod *v1.Pod, container *v1.Container) error

	// UpdatePluginResources updates node resources based on devices already
	// allocated to pods. The node object is provided for the device manager to
	// update the node capacity to reflect the currently available devices.
	UpdatePluginResources(node *schedulerframework.NodeInfo, attrs *lifecycle.PodAdmitAttributes) error

	// Stop stops the manager.
	Stop() error

	// GetDeviceRunContainerOptions checks whether we have cached containerDevices
	// for the passed-in <pod, container> and returns its DeviceRunContainerOptions
	// for the found one. An empty struct is returned in case no cached state is found.
	GetDeviceRunContainerOptions(pod *v1.Pod, container *v1.Container) (*DeviceRunContainerOptions, error)

	// GetCapacity returns the amount of available device plugin resource capacity, resource allocatable
	// and inactive device plugin resources previously registered on the node.
	GetCapacity() (v1.ResourceList, v1.ResourceList, []string)
	GetWatcherHandler() cache.PluginHandler

	// GetDevices returns information about the devices assigned to pods and containers
	GetDevices(podUID, containerName string) []*podresourcesapi.ContainerDevices

	// ShouldResetExtendedResourceCapacity returns whether the extended resources should be reset or not,
	// depending on the checkpoint file availability. Absence of the checkpoint file strongly indicates
	// the node has been recreated.
	ShouldResetExtendedResourceCapacity() bool

	// TopologyManager HintProvider provider indicates the Device Manager implements the Topology Manager Interface
	// and is consulted to make Topology aware resource alignments
	GetTopologyHints(pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint

	// UpdateAllocatedDevices frees any Devices that are bound to terminated pods.
	UpdateAllocatedDevices()
}

// DeviceRunContainerOptions contains the combined container runtime settings to consume its allocated devices.
type DeviceRunContainerOptions struct {
	// The environment variables list.
	Envs []kubecontainer.EnvVar
	// The mounts for the container.
	Mounts []kubecontainer.Mount
	// The host devices mapped into the container.
	Devices []kubecontainer.DeviceInfo
	// The Annotations for the container
	Annotations []kubecontainer.Annotation
}

// kubeletDeviceManagerCheckpoint is the file name of device plugin checkpoint
const kubeletDeviceManagerCheckpoint = "kubelet_internal_checkpoint"
