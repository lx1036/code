package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	componentbaseconfig "k8s.io/component-base/config"
	"time"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DeschedulerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// Time interval for descheduler to run
	DeschedulingInterval time.Duration `json:"deschedulingInterval,omitempty"`

	// KubeconfigFile is path to kubeconfig file with authorization and master
	// location information.
	KubeconfigFile string `json:"kubeconfigFile"`

	// PolicyConfigFile is the filepath to the descheduler policy configuration.
	PolicyConfigFile string `json:"policyConfigFile,omitempty"`

	// Dry run
	DryRun bool `json:"dryRun,omitempty"`

	// Node selectors
	NodeSelector string `json:"nodeSelector,omitempty"`

	// MaxNoOfPodsToEvictPerNode restricts maximum of pods to be evicted per node.
	MaxNoOfPodsToEvictPerNode int `json:"maxNoOfPodsToEvictPerNode,omitempty"`

	// EvictLocalStoragePods allows pods using local storage to be evicted.
	EvictLocalStoragePods bool `json:"evictLocalStoragePods,omitempty"`

	// IgnorePVCPods sets whether PVC pods should be allowed to be evicted
	IgnorePVCPods bool `json:"ignorePvcPods,omitempty"`

	// Logging specifies the options of logging.
	// Refer [Logs Options](https://github.com/kubernetes/component-base/blob/master/logs/options.go) for more information.
	Logging componentbaseconfig.LoggingConfiguration `json:"logging,omitempty"`
}
