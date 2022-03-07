package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Eip `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=eip,singular=eip
// +kubebuilder:printcolumn:name="Address",type="string",JSONPath=".spec.address"
// +kubebuilder:printcolumn:name="Protocol",type="string",JSONPath=".spec.protocol"
// +kubebuilder:printcolumn:name="Interface",type="string",JSONPath=".spec.interface"

type Eip struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EipSpec   `json:"spec,omitempty"`
	Status EipStatus `json:"status,omitempty"`
}

// EipSpec defines the desired state of EIP
type EipSpec struct {
	// +kubebuilder:validation:Required
	Address string `json:"address,required"`
	// +kubebuilder:validation:Enum=bgp;layer2
	Protocol      string `json:"protocol,omitempty"`
	Interface     string `json:"interface,omitempty"`
	Disable       bool   `json:"disable,omitempty"`
	UsingKnownIPs bool   `json:"usingKnownIPs,omitempty"`
}

// EipStatus defines the observed state of EIP
type EipStatus struct {
	Occupied bool              `json:"occupied,omitempty"`
	Usage    int               `json:"usage,omitempty"`
	PoolSize int               `json:"poolSize,omitempty"`
	Used     map[string]string `json:"used,omitempty"`
	FirstIP  string            `json:"firstIP,omitempty"`
	LastIP   string            `json:"lastIP,omitempty"`
	Ready    bool              `json:"ready,omitempty"`
	V4       bool              `json:"v4,omitempty"`
}
