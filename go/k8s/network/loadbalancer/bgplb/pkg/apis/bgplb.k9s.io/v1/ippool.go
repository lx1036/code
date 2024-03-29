package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type IPPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPPool `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ipp,singular=ippool
// +kubebuilder:printcolumn:name="Cidr",type="string",JSONPath=".spec.cidr"
// +kubebuilder:printcolumn:name="PoolSize",type="string",JSONPath=".status.poolSize"
// +kubebuilder:printcolumn:name="Usage",type="string",JSONPath=".status.usage"

type IPPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPPoolSpec   `json:"spec,omitempty"`
	Status IPPoolStatus `json:"status,omitempty"`
}

type IPPoolSpec struct {
	// +kubebuilder:validation:Required
	Cidr string `json:"cidr,required"`
}

type IPPoolStatus struct {
	PoolSize int    `json:"poolSize,omitempty"`
	FirstIP  string `json:"firstIP,omitempty"`
	LastIP   string `json:"lastIP,omitempty"`

	Usage int               `json:"usage,omitempty"`
	Used  map[string]string `json:"used,omitempty"`
}
