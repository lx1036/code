package v1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,singular=podgroup
// +kubebuilder:printcolumn:name="Queue",type="string",JSONPath=".spec.queue"

// PodGroup is a collection of Pod; used for batch workload.
type PodGroup struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod group.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec PodGroupSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the current information about a pod group.
	// This data may not be up to date.
	// +optional
	Status PodGroupStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodGroupList is a collection of pod groups.
type PodGroupList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of PodGroup
	Items []PodGroup `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// PodGroupSpec represents the template of a pod group.
type PodGroupSpec struct {
	// MinMember defines the minimal number of members/tasks to run the pod group;
	// if there's not enough resources to start all tasks, the scheduler
	// will not start anyone.
	MinMember int32 `json:"minMember,omitempty" protobuf:"bytes,1,opt,name=minMember"`

	// Queue defines the queue to allocate resource for PodGroup; if queue does not exist,
	// the PodGroup will not be scheduled.
	Queue string `json:"queue,omitempty" protobuf:"bytes,2,opt,name=queue"`

	// If specified, indicates the PodGroup's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the PodGroup priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,3,opt,name=priorityClassName"`
}

// PodGroupPhase is the phase of a pod group at the current time.
type PodGroupPhase string

// These are the valid phase of podGroups.
const (
	// PodPending means the pod group has been accepted by the system, but scheduler can not allocate
	// enough resources to it.
	PodGroupPending PodGroupPhase = "Pending"

	// PodRunning means `spec.minMember` pods of PodGroups has been in running phase.
	PodGroupRunning PodGroupPhase = "Running"

	// PodGroupUnknown means part of `spec.minMember` pods are running but the other part can not
	// be scheduled, e.g. not enough resource; scheduler will wait for related controller to recover it.
	PodGroupUnknown PodGroupPhase = "Unknown"
)

type PodGroupConditionType string

const (
	PodGroupUnschedulableType PodGroupConditionType = "Unschedulable"
)

// PodGroupCondition contains details for the current state of this pod group.
type PodGroupCondition struct {
	// Type is the type of the condition
	Type PodGroupConditionType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type"`

	// Status is the status of the condition.
	Status v1.ConditionStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status"`

	// The ID of condition transition.
	TransitionID string `json:"transitionID,omitempty" protobuf:"bytes,3,opt,name=transitionID"`

	// Last time the phase transitioned from another to current phase.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`

	// Unique, one-word, CamelCase reason for the phase's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// PodGroupStatus represents the current state of a pod group.
type PodGroupStatus struct {
	// Current phase of PodGroup.
	Phase PodGroupPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase"`

	// The conditions of PodGroup.
	// +optional
	Conditions []PodGroupCondition `json:"conditions,omitempty" protobuf:"bytes,2,opt,name=conditions"`

	// The number of actively running pods.
	// +optional
	Running int32 `json:"running,omitempty" protobuf:"bytes,3,opt,name=running"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded,omitempty" protobuf:"bytes,4,opt,name=succeeded"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed,omitempty" protobuf:"bytes,5,opt,name=failed"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:singular=queue

// Queue is a queue of PodGroup.
type Queue struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the queue.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status
	// +optional
	Spec QueueSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// The status of queue.
	// +optional
	Status QueueStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// QueueList is a collection of queues.
type QueueList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of PodGroup
	Items []Queue `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// QueueSpec represents the template of Queue.
type QueueSpec struct {
	Weight     int32           `json:"weight,omitempty" protobuf:"bytes,1,opt,name=weight"`
	Capability v1.ResourceList `json:"capability,omitempty" protobuf:"bytes,2,opt,name=capability"`
}

// QueueStatus represents the status of Queue.
type QueueStatus struct {
	// The number of 'Unknonw' PodGroup in this queue.
	Unknown int32 `json:"unknown,omitempty" protobuf:"bytes,1,opt,name=unknown"`
	// The number of 'Pending' PodGroup in this queue.
	Pending int32 `json:"pending,omitempty" protobuf:"bytes,2,opt,name=pending"`
	// The number of 'Running' PodGroup in this queue.
	Running int32 `json:"running,omitempty" protobuf:"bytes,3,opt,name=running"`
}
