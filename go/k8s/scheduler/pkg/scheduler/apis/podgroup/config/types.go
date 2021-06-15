package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CoschedulingArgs defines the parameters for Coscheduling plugin.
type CoschedulingArgs struct {
	metav1.TypeMeta
	
	// PermitWaitingTime is the wait timeout in seconds.
	PermitWaitingTimeSeconds int64
	// DeniedPGExpirationTimeSeconds is the expiration time of the denied podgroup store.
	DeniedPGExpirationTimeSeconds int64
	// KubeMaster is the url of api-server
	KubeMaster string
	// KubeConfigPath for scheduler
	KubeConfigPath string
}
