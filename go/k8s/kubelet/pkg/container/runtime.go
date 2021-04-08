package container

import (
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/flowcontrol"
)

// Runtime interface defines the interfaces that should be implemented
// by a container runtime.
// Thread safety is required from implementations of this interface.
type Runtime interface {
	// Type returns the type of the container runtime.
	Type() string

	//SupportsSingleFileMapping returns whether the container runtime supports single file mappings or not.
	SupportsSingleFileMapping() bool

	// Version returns the version information of the container runtime.
	Version() (Version, error)

	// APIVersion returns the cached API version information of the container
	// runtime. Implementation is expected to update this cache periodically.
	// This may be different from the runtime engine's version.
	// TODO(random-liu): We should fold this into Version()
	APIVersion() (Version, error)
	// Status returns the status of the runtime. An error is returned if the Status
	// function itself fails, nil otherwise.
	Status() (*RuntimeStatus, error)
	// GetPods returns a list of containers grouped by pods. The boolean parameter
	// specifies whether the runtime returns all containers including those already
	// exited and dead containers (used for garbage collection).
	GetPods(all bool) ([]*Pod, error)
	// GarbageCollect removes dead containers using the specified container gc policy
	// If allSourcesReady is not true, it means that kubelet doesn't have the
	// complete list of pods from all available sources (e.g., apiserver, http,
	// file). In this case, garbage collector should refrain itself from aggressive
	// behavior such as removing all containers of unrecognized pods (yet).
	// If evictNonDeletedPods is set to true, containers and sandboxes belonging to pods
	// that are terminated, but not deleted will be evicted.  Otherwise, only deleted pods will be GC'd.
	// INFO: Revisit this method and make it cleaner.
	GarbageCollect(gcPolicy GCPolicy, allSourcesReady bool, evictNonDeletedPods bool) error
	// Syncs the running pod into the desired pod.
	SyncPod(pod *v1.Pod, podStatus *PodStatus, pullSecrets []v1.Secret, backOff *flowcontrol.Backoff) PodSyncResult
	// KillPod kills all the containers of a pod. Pod may be nil, running pod must not be.
	// TODO(random-liu): Return PodSyncResult in KillPod.
	// gracePeriodOverride if specified allows the caller to override the pod default grace period.
	// only hard kill paths are allowed to specify a gracePeriodOverride in the kubelet in order to not corrupt user data.
	// it is useful when doing SIGKILL for hard eviction scenarios, or max grace period during soft eviction scenarios.
	KillPod(pod *v1.Pod, runningPod Pod, gracePeriodOverride *int64) error
	// GetPodStatus retrieves the status of the pod, including the
	// information of all containers in the pod that are visible in Runtime.
	GetPodStatus(uid types.UID, name, namespace string) (*PodStatus, error)
	// TODO(vmarmol): Unify pod and containerID args.
	// GetContainerLogs returns logs of a specific container. By
	// default, it returns a snapshot of the container log. Set 'follow' to true to
	// stream the log. Set 'follow' to false and specify the number of lines (e.g.
	// "100" or "all") to tail the log.
	GetContainerLogs(ctx context.Context, pod *v1.Pod, containerID ContainerID, logOptions *v1.PodLogOptions, stdout, stderr io.Writer) (err error)
	// Delete a container. If the container is still running, an error is returned.
	DeleteContainer(containerID ContainerID) error
	// ImageService provides methods to image-related methods.
	ImageService
	// UpdatePodCIDR sends a new podCIDR to the runtime.
	// This method just proxies a new runtimeConfig with the updated
	// CIDR value down to the runtime shim.
	UpdatePodCIDR(podCIDR string) error
}
