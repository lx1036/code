package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CapacitySchedulingArgs defines the scheduling parameters for CapacityScheduling plugin.
type CapacitySchedulingArgs struct {
	metav1.TypeMeta

	// KubeConfigPath is the path of kubeconfig.
	KubeConfigPath string
}
