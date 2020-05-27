package resourcequota

import (
	"k8s-lx1036/k8s-ui/dashboard/controllers/resource/common"
	corev1 "k8s.io/api/core/v1"
)

type ResourceQuotaDetailList struct {
	ListMeta common.ListMeta `json:"listMeta"`
	Items []ResourceQuotaDetail `json:"items"`
}

type ResourceQuotaDetail struct {
	ObjectMeta common.ObjectMeta `json:"objectMeta"`
	TypeMeta   common.TypeMeta   `json:"typeMeta"`
	
	Scopes []corev1.ResourceQuotaScope `json:"scopes"`
	StatusList map[corev1.ResourceName]ResourceStatus `json:"statusList"`
}

type ResourceStatus struct {
	Used string `json:"used,omitempty"`
	Hard string `json:"hard,omitempty"`
}

func ToResourceQuotaDetail(rawResourceQuota *corev1.ResourceQuota) ResourceQuotaDetail {
	statusList := map[corev1.ResourceName]ResourceStatus{}
	for key, value := range rawResourceQuota.Status.Hard {
		used := rawResourceQuota.Status.Used[key]
		statusList[key] = ResourceStatus{
			Used: used.String(),
			Hard: value.String(),
		}
	}
	
	return ResourceQuotaDetail{
		ObjectMeta: common.NewObjectMeta(rawResourceQuota.ObjectMeta),
		TypeMeta:   common.NewTypeMeta(common.ResourceKindResourceQuota),
		Scopes:     rawResourceQuota.Spec.Scopes,
		StatusList: statusList,
	}
}
