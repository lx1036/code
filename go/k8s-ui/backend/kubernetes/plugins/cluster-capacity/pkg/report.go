package pkg

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type ClusterCapacityReview struct {
	metav1.TypeMeta
	Spec   ClusterCapacityReviewSpec   `json:"spec"`
	Status ClusterCapacityReviewStatus `json:"status"`
}
type ClusterCapacityReviewSpec struct {
	// the pod desired for scheduling
	Templates []corev1.Pod `json:"templates"`

	// desired number of replicas that should be scheduled
	// +optional
	Replicas int32 `json:"replicas"`

	PodRequirements []*Requirements `json:"podRequirements"`
}
type Requirements struct {
	PodName       string            `json:"podName"`
	Resources     *Resources        `json:"resources"`
	NodeSelectors map[string]string `json:"nodeSelectors"`
}
type Resources struct {
	PrimaryResources corev1.ResourceList           `json:"primaryResources"`
	ScalarResources  map[corev1.ResourceName]int64 `json:"scalarResources"`
}
type ClusterCapacityReviewResult struct {
	PodName string `json:"podName"`
	// numbers of replicas on nodes
	ReplicasOnNodes []*ReplicasOnNode `json:"replicasOnNodes"`
	// reason why no more pods could schedule (if any on this node)
	FailSummary []FailReasonSummary `json:"failSummary"`
}
type ReplicasOnNode struct {
	NodeName string `json:"nodeName"`
	Replicas int    `json:"replicas"`
}
type FailReasonSummary struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}
type ClusterCapacityReviewStatus struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	// actual number of replicas that could schedule
	Replicas int32 `json:"replicas"`

	FailReason *ClusterCapacityReviewScheduleFailReason `json:"failReason"`

	// per node information about the scheduling simulation
	Pods []*ClusterCapacityReviewResult `json:"pods"`
}
type ClusterCapacityReviewScheduleFailReason struct {
	FailType    string `json:"failType"`
	FailMessage string `json:"failMessage"`
}

func clusterCapacityReviewPrintJson(r *ClusterCapacityReview) error {
	jsoned, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("failed to create json: %v", err)
	}
	fmt.Println(string(jsoned))
	return nil
}
func clusterCapacityReviewPrintYaml(r *ClusterCapacityReview) error {
	yamled, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("Failed to create yaml: %v", err)
	}
	fmt.Print(string(yamled))
	return nil
}

func ClusterCapacityReviewPrint(r *ClusterCapacityReview, verbose bool, format string) error {
	switch format {
	case "json":
		return clusterCapacityReviewPrintJson(r)
	case "yaml":
		return clusterCapacityReviewPrintYaml(r)
	default:
		return fmt.Errorf("output format %q not recognized", format)
	}
}

func GetReport(pods []*corev1.Pod, status Status) *ClusterCapacityReview {
	return &ClusterCapacityReview{
		Spec:   getReviewSpec(pods),
		Status: getReviewStatus(pods, status),
	}
}
