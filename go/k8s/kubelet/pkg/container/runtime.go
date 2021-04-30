package container

import (
	"context"
	"io"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/flowcontrol"
	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

// Version interface allow to consume the runtime versions - compare and format to string.
type Version interface {
	// Compare compares two versions of the runtime. On success it returns -1
	// if the version is less than the other, 1 if it is greater than the other,
	// or 0 if they are equal.
	Compare(other string) (int, error)
	// String returns a string that represents the version.
	String() string
}

// RuntimeStatus contains the status of the runtime.
type RuntimeStatus struct {
	// Conditions is an array of current observed runtime conditions.
	Conditions []RuntimeCondition
}

// RuntimeConditionType is the types of required runtime conditions.
type RuntimeConditionType string

const (
	// RuntimeReady means the runtime is up and ready to accept basic containers.
	RuntimeReady RuntimeConditionType = "RuntimeReady"
	// NetworkReady means the runtime network is up and ready to accept containers which require network.
	NetworkReady RuntimeConditionType = "NetworkReady"
)

// RuntimeCondition contains condition information for the runtime.
type RuntimeCondition struct {
	// Type of runtime condition.
	Type RuntimeConditionType
	// Status of the condition, one of true/false.
	Status bool
	// Reason is brief reason for the condition's last transition.
	Reason string
	// Message is human readable message indicating details about last transition.
	Message string
}

// Pod is a group of containers.
type Pod struct {
	// The ID of the pod, which can be used to retrieve a particular pod
	// from the pod list returned by GetPods().
	ID types.UID
	// The name and namespace of the pod, which is readable by human.
	Name      string
	Namespace string
	// List of containers that belongs to this pod. It may contain only
	// running containers, or mixed with dead ones (when GetPods(true)).
	Containers []*Container
	// List of sandboxes associated with this pod. The sandboxes are converted
	// to Container temporarily to avoid substantial changes to other
	// components. This is only populated by kuberuntime.
	// TODO: use the runtimeApi.PodSandbox type directly.
	Sandboxes []*Container
}

// PodStatus represents the status of the pod and its containers.
// v1.PodStatus can be derived from examining PodStatus and v1.Pod.
type PodStatus struct {
	// ID of the pod.
	ID types.UID
	// Name of the pod.
	Name string
	// Namespace of the pod.
	Namespace string
	// All IPs assigned to this pod
	IPs []string
	// Status of containers in the pod.
	ContainerStatuses []*Status
	// Status of the pod sandbox.
	// Only for kuberuntime now, other runtime may keep it nil.
	SandboxStatuses []*runtimeapi.PodSandboxStatus
}

// Status represents the status of a container.
type Status struct {
	// ID of the container.
	ID ContainerID
	// Name of the container.
	Name string
	// Status of the container.
	State State
	// Creation time of the container.
	CreatedAt time.Time
	// Start time of the container.
	StartedAt time.Time
	// Finish time of the container.
	FinishedAt time.Time
	// Exit code of the container.
	ExitCode int
	// Name of the image, this also includes the tag of the image,
	// the expected form is "NAME:TAG".
	Image string
	// ID of the image.
	ImageID string
	// Hash of the container, used for comparison.
	Hash uint64
	// Number of times that the container has been restarted.
	RestartCount int
	// A string explains why container is in such a status.
	Reason string
	// Message written by the container before exiting (stored in
	// TerminationMessagePath).
	Message string
}

// State represents the state of a container
type State string

const (
	// ContainerStateCreated indicates a container that has been created (e.g. with docker create) but not started.
	ContainerStateCreated State = "created"
	// ContainerStateRunning indicates a currently running container.
	ContainerStateRunning State = "running"
	// ContainerStateExited indicates a container that ran and completed ("stopped" in other contexts, although a created container is technically also "stopped").
	ContainerStateExited State = "exited"
	// ContainerStateUnknown encompasses all the states that we currently don't care about (like restarting, paused, dead).
	ContainerStateUnknown State = "unknown"
)

// Container provides the runtime information for a container, such as ID, hash,
// state of the container.
type Container struct {
	// The ID of the container, used by the container runtime to identify
	// a container.
	ID ContainerID
	// The name of the container, which should be the same as specified by
	// v1.Container.
	Name string
	// The image name of the container, this also includes the tag of the image,
	// the expected form is "NAME:TAG".
	Image string
	// The id of the image used by the container.
	ImageID string
	// Hash of the container, used for comparison. Optional for containers
	// not managed by kubelet.
	Hash uint64
	// State is the state of the container.
	State State
}

// ContainerID is a type that identifies a container.
type ContainerID struct {
	// The type of the container runtime. e.g. 'docker'.
	Type string
	// The identification of the container, this is comsumable by
	// the underlying container runtime. (Note that the container
	// runtime interface still takes the whole struct as input).
	ID string
}

// Annotation represents an annotation.
type Annotation struct {
	Name  string
	Value string
}

// ImageSpec is an internal representation of an image.  Currently, it wraps the
// value of a Container's Image field, but in the future it will include more detailed
// information about the different image types.
type ImageSpec struct {
	// ID of the image.
	Image string
	// The annotations for the image.
	// This should be passed to CRI during image pulls and returned when images are listed.
	Annotations []Annotation
}

// ImageService interfaces allows to work with image service.
type ImageService interface {
	// PullImage pulls an image from the network to local storage using the supplied
	// secrets if necessary. It returns a reference (digest or ID) to the pulled image.
	PullImage(image ImageSpec, pullSecrets []v1.Secret, podSandboxConfig *runtimeapi.PodSandboxConfig) (string, error)
	// GetImageRef gets the reference (digest or ID) of the image which has already been in
	// the local storage. It returns ("", nil) if the image isn't in the local storage.
	GetImageRef(image ImageSpec) (string, error)
	// Gets all images currently on the machine.
	ListImages() ([]Image, error)
	// Removes the specified image.
	RemoveImage(image ImageSpec) error
	// Returns Image statistics.
	ImageStats() (*ImageStats, error)
}

// Image contains basic information about a container image.
type Image struct {
	// ID of the image.
	ID string
	// Other names by which this image is known.
	RepoTags []string
	// Digests by which this image is known.
	RepoDigests []string
	// The size of the image in bytes.
	Size int64
	// ImageSpec for the image which include annotations.
	Spec ImageSpec
}

// ImageStats contains statistics about all the images currently available.
type ImageStats struct {
	// Total amount of storage consumed by existing images.
	TotalStorageBytes uint64
}

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
