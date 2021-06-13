package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
)

// @see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/generating-clientset.md
// 额外的打印列: https://kubernetes.io/zh/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#additional-printer-columns
// https://cloudnative.to/kubebuilder/reference/markers/crd.html

/*
kubectl get vpa
NAME          MODE   CPU    MEM       PROVIDED   AGE
hamster-vpa   Auto   587m   262144k   True       8d
*/

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Mode",type="string",JSONPath=".spec.updatePolicy.updateMode"
// +kubebuilder:printcolumn:name="CPU",type="string",JSONPath=".status.recommendation.containerRecommendations[0].target.cpu"
// +kubebuilder:printcolumn:name="Mem",type="string",JSONPath=".status.recommendation.containerRecommendations[0].target.memory"
// +kubebuilder:printcolumn:name="Provided",type="string",JSONPath=".status.conditions[?(@.type=='RecommendationProvided')].status"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type VerticalPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec VerticalPodAutoscalerSpec `json:"spec" protobuf:"bytes,2,name=spec"`

	// +optional
	Status VerticalPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

/*
spec:
  resourcePolicy:
    containerPolicies:
    - containerName: '*'
      controlledResources:
      - cpu
      - memory
      maxAllowed:
        cpu: 1
        memory: 500Mi
      minAllowed:
        cpu: 100m
        memory: 50Mi
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: hamster
  updatePolicy:
    updateMode: Auto
*/

type VerticalPodAutoscalerSpec struct {
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef" protobuf:"bytes,1,name=targetRef"`

	// +optional
	UpdatePolicy *PodUpdatePolicy `json:"updatePolicy,omitempty" protobuf:"bytes,2,opt,name=updatePolicy"`

	// +optional
	ResourcePolicy *PodResourcePolicy `json:"resourcePolicy,omitempty" protobuf:"bytes,3,opt,name=resourcePolicy"`
}

type PodUpdatePolicy struct {
	// Controls when autoscaler applies changes to the pod resources.
	// The default is 'Auto'.
	// +optional
	UpdateMode *UpdateMode `json:"updateMode,omitempty" protobuf:"bytes,1,opt,name=updateMode"`
}

/*
  conditions:
  - lastTransitionTime: "2021-06-05T08:22:55Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: hamster
      lowerBound:
        cpu: 586m
        memory: 262144k
      target:
        cpu: 587m
        memory: 262144k
      uncappedTarget:
        cpu: 587m
        memory: 262144k
      upperBound:
        cpu: 967m
        memory: 262144k
*/

type VerticalPodAutoscalerStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerticalPodAutoscalerList is a list of VerticalPodAutoscaler objects.
type VerticalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of vertical pod autoscaler objects.
	Items []VerticalPodAutoscaler `json:"items" protobuf:"bytes,2,rep,name=items"`
}

// VerticalPodAutoscalerConditionType are the valid conditions of
// a VerticalPodAutoscaler.
type VerticalPodAutoscalerConditionType string

var (
	// RecommendationProvided indicates whether the VPA recommender was able to calculate a recommendation.
	RecommendationProvided VerticalPodAutoscalerConditionType = "RecommendationProvided"
	// LowConfidence indicates whether the VPA recommender has low confidence in the recommendation for
	// some of containers.
	LowConfidence VerticalPodAutoscalerConditionType = "LowConfidence"
	// NoPodsMatched indicates that label selector used with VPA object didn't match any pods.
	NoPodsMatched VerticalPodAutoscalerConditionType = "NoPodsMatched"
	// FetchingHistory indicates that VPA recommender is in the process of loading additional history samples.
	FetchingHistory VerticalPodAutoscalerConditionType = "FetchingHistory"
	// ConfigDeprecated indicates that this VPA configuration is deprecated
	// and will stop being supported soon.
	ConfigDeprecated VerticalPodAutoscalerConditionType = "ConfigDeprecated"
	// ConfigUnsupported indicates that this VPA configuration is unsupported
	// and recommendations will not be provided for it.
	ConfigUnsupported VerticalPodAutoscalerConditionType = "ConfigUnsupported"
)

// RecommendedPodResources is the recommendation of resources computed by
// autoscaler. It contains a recommendation for each container in the pod
// (except for those with `ContainerScalingMode` set to 'Off').
type RecommendedPodResources struct {
	// Resources recommended by the autoscaler for each container.
	// +optional
	ContainerRecommendations []RecommendedContainerResources `json:"containerRecommendations,omitempty" protobuf:"bytes,1,rep,name=containerRecommendations"`
}

// RecommendedContainerResources is the recommendation of resources computed by
// autoscaler for a specific container. Respects the container resource policy
// if present in the spec. In particular the recommendation is not produced for
// containers with `ContainerScalingMode` set to 'Off'.
type RecommendedContainerResources struct {
	// Name of the container.
	ContainerName string `json:"containerName,omitempty" protobuf:"bytes,1,opt,name=containerName"`
	// Recommended amount of resources. Observes ContainerResourcePolicy.
	Target v1.ResourceList `json:"target" protobuf:"bytes,2,rep,name=target,casttype=ResourceList,castkey=ResourceName"`
	// Minimum recommended amount of resources. Observes ContainerResourcePolicy.
	// This amount is not guaranteed to be sufficient for the application to operate in a stable way, however
	// running with less resources is likely to have significant impact on performance/availability.
	// +optional
	LowerBound v1.ResourceList `json:"lowerBound,omitempty" protobuf:"bytes,3,rep,name=lowerBound,casttype=ResourceList,castkey=ResourceName"`
	// Maximum recommended amount of resources. Observes ContainerResourcePolicy.
	// Any resources allocated beyond this value are likely wasted. This value may be larger than the maximum
	// amount of application is actually capable of consuming.
	// +optional
	UpperBound v1.ResourceList `json:"upperBound,omitempty" protobuf:"bytes,4,rep,name=upperBound,casttype=ResourceList,castkey=ResourceName"`
	// The most recent recommended resources target computed by the autoscaler
	// for the controlled pods, based only on actual resource usage, not taking
	// into account the ContainerResourcePolicy.
	// May differ from the Recommendation if the actual resource usage causes
	// the target to violate the ContainerResourcePolicy (lower than MinAllowed
	// or higher that MaxAllowed).
	// Used only as status indication, will not affect actual resource assignment.
	// +optional
	UncappedTarget v1.ResourceList `json:"uncappedTarget,omitempty" protobuf:"bytes,5,opt,name=uncappedTarget"`
}

// UpdateMode controls when autoscaler applies changes to the pod resoures.
// +kubebuilder:validation:Enum=Off;Initial;Recreate;Auto
type UpdateMode string

const (
	// UpdateModeOff means that autoscaler never changes Pod resources.
	// The recommender still sets the recommended resources in the
	// VerticalPodAutoscaler object. This can be used for a "dry run".
	UpdateModeOff UpdateMode = "Off"
	// UpdateModeInitial means that autoscaler only assigns resources on pod
	// creation and does not change them during the lifetime of the pod.
	UpdateModeInitial UpdateMode = "Initial"
	// UpdateModeRecreate means that autoscaler assigns resources on pod
	// creation and additionally can update them during the lifetime of the
	// pod by deleting and recreating the pod.
	UpdateModeRecreate UpdateMode = "Recreate"
	// UpdateModeAuto means that autoscaler assigns resources on pod creation
	// and additionally can update them during the lifetime of the pod,
	// using any available update method. Currently this is equivalent to
	// Recreate, which is the only available update method.
	UpdateModeAuto UpdateMode = "Auto"
)

// ContainerScalingMode controls whether autoscaler is enabled for a specific
// container.
// +kubebuilder:validation:Enum=Auto;Off
type ContainerScalingMode string

const (
	// ContainerScalingModeAuto means autoscaling is enabled for a container.
	ContainerScalingModeAuto ContainerScalingMode = "Auto"
	// ContainerScalingModeOff means autoscaling is disabled for a container.
	ContainerScalingModeOff ContainerScalingMode = "Off"
)
