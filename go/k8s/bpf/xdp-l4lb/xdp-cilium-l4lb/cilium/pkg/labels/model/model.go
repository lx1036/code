package model

import (
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
)

func NewModel(o *labels.OpLabels) *models.LabelConfigurationStatus {
	return &models.LabelConfigurationStatus{
		Realized: &models.LabelConfigurationSpec{
			User: o.Custom.GetModel(),
		},
		SecurityRelevant: o.OrchestrationIdentity.GetModel(),
		Derived:          o.OrchestrationInfo.GetModel(),
		Disabled:         o.Disabled.GetModel(),
	}
}
