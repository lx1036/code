package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/autoscaling"
)

// @see https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/generating-clientset.md

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type VerticalPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec VerticalPodAutoscalerSpec `json:"spec" protobuf:"bytes,2,name=spec"`

	Status VerticalPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

type VerticalPodAutoscalerSpec struct {
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef" protobuf:"bytes,1,name=targetRef"`
}

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
