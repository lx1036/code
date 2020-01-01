package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ResourceName = string

const (
	ResourceNameDeployment  ResourceName = "deployments"
	ResourceNamePod         ResourceName = "pods"
	ResourceNameCronJob     ResourceName = "cronjobs"
	ResourceNameDaemonSet   ResourceName = "daemonsets"
	ResourceNameStatefulSet ResourceName = "statefulsets"
	ResourceNameJob         ResourceName = "jobs"
)

type KindName = string

const (
	KindNamePod KindName = "Pod"
)

type ResourceMap struct {
	GroupVersionResourceKind GroupVersionResourceKind
	Namespaced               bool
}

type GroupVersionResourceKind struct {
	schema.GroupVersionResource
	Kind string
}

var KindToResourceMap = map[string]ResourceMap{
	ResourceNamePod: {
		GroupVersionResourceKind: GroupVersionResourceKind{
			GroupVersionResource: schema.GroupVersionResource{
				Group:    corev1.GroupName,
				Version:  corev1.SchemeGroupVersion.Version,
				Resource: ResourceNamePod,
			},
			Kind: KindNamePod,
		},
		Namespaced: true,
	},
}
