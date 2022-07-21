package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IPPoolList contains a list of IPPool
type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPPool `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ipp,singular=ippool
// +kubebuilder:printcolumn:name="Range",type="string",JSONPath=".spec.range"

type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPPoolSpec   `json:"spec,omitempty"`
	Status IPPoolStatus `json:"status,omitempty"`
}

func (ippool *IPPool) Allocations() {

}

type IPPoolSpec struct {
	// Range is a RFC 4632/4291-style string that represents an IP address and prefix length in CIDR notation
	Range string `json:"range"`
	// Allocations is the set of allocated IPs for the given range. Its` indices are a direct mapping to the
	// IP with the same index/offset for the pool's range.
	Allocations map[string]IPAllocation `json:"allocations,omitempty"`
}

// IPAllocation represents metadata about the pod/container owner of a specific IP
type IPAllocation struct {
	ContainerID string `json:"id"`
	PodRef      string `json:"podref,omitempty"`
}

type IPPoolStatus struct {
	PoolSize int    `json:"poolSize,omitempty"`
	FirstIP  string `json:"firstIP,omitempty"`
	LastIP   string `json:"lastIP,omitempty"`

	Usage int               `json:"usage,omitempty"`
	Used  map[string]string `json:"used,omitempty"`
}
