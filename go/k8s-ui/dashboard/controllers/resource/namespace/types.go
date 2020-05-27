package namespace

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/limitrange"
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common/resourcequota"
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

type NamespaceDetail struct {
	Namespace Namespace `json:"namespace"`
	
	ResourceQuotaList *resourcequota.ResourceQuotaDetailList `json:"resourceQuotaList"`
	ResourceLimits []limitrange.LimitRangeItem `json:"resourceLimits"`
	
	Errors []error `json:"errors"`
}
