package namespace

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	corev1 "k8s.io/api/core/v1"
)

type Namespace struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta common.TypeMeta `json:"typeMeta"`
	
	Phase corev1.NamespacePhase `json:"phase"`
}

type NamespaceList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	
	Namespaces []Namespace `json:"namespaces"`
	
	Errors []error `json:"errors"`
}

