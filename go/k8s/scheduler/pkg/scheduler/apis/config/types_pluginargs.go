package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeLabelArgs holds arguments used to configure the NodeLabel plugin.
type NodeLabelArgs struct {
	metav1.TypeMeta

	// PresentLabels should be present for the node to be considered a fit for hosting the pod
	PresentLabels []string
	// AbsentLabels should be absent for the node to be considered a fit for hosting the pod
	AbsentLabels []string
	// Nodes that have labels in the list will get a higher score.
	PresentLabelsPreference []string
	// Nodes that don't have labels in the list will get a higher score.
	AbsentLabelsPreference []string
}
