package common

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
)


// List of all resource kinds supported by the UI.
const (
	ResourceKindConfigMap                = "configmap"
	ResourceKindDaemonSet                = "daemonset"
	ResourceKindDeployment               = "deployment"
	ResourceKindEvent                    = "event"
	ResourceKindHorizontalPodAutoscaler  = "horizontalpodautoscaler"
	ResourceKindIngress                  = "ingress"
	ResourceKindJob                      = "job"
	ResourceKindCronJob                  = "cronjob"
	ResourceKindLimitRange               = "limitrange"
	ResourceKindNamespace                = "namespace"
	ResourceKindNode                     = "node"
	ResourceKindPersistentVolumeClaim    = "persistentvolumeclaim"
	ResourceKindPersistentVolume         = "persistentvolume"
	ResourceKindCustomResourceDefinition = "customresourcedefinition"
	ResourceKindPod                      = "pod"
	ResourceKindReplicaSet               = "replicaset"
	ResourceKindReplicationController    = "replicationcontroller"
	ResourceKindResourceQuota            = "resourcequota"
	ResourceKindSecret                   = "secret"
	ResourceKindService                  = "service"
	ResourceKindStatefulSet              = "statefulset"
	ResourceKindStorageClass             = "storageclass"
	ResourceKindClusterRole              = "clusterrole"
	ResourceKindClusterRoleBinding       = "clusterrolebinding"
	ResourceKindRole                     = "role"
	ResourceKindRoleBinding              = "rolebinding"
	ResourceKindPlugin                   = "plugin"
	ResourceKindEndpoint                 = "endpoint"
)

var ListEverything = metav1.ListOptions{
	LabelSelector: labels.Everything().String(),
	FieldSelector: fields.Everything().String(),
}
var GetEverything = metav1.GetOptions{}
var CreateEverything = metav1.CreateOptions{}

type ResourceStatus struct {
	// Number of resources that are currently in running state.
	Running int `json:"running"`
	
	// Number of resources that are currently in pending state.
	Pending int `json:"pending"`
	
	// Number of resources that are in failed state.
	Failed int `json:"failed"`
	
	// Number of resources that are in succeeded state.
	Succeeded int `json:"succeeded"`
}

// ListMeta describes list of objects, i.e. holds information about pagination options set for the list.
type ListMeta struct {
	// Total number of items on the list. Used for pagination.
	TotalItems int `json:"totalItems"`
}

type ResourceKind string
func (k ResourceKind) Scalable() bool {
	scalable := []ResourceKind{
		ResourceKindDeployment,
		ResourceKindReplicaSet,
		ResourceKindReplicationController,
		ResourceKindStatefulSet,
	}

	for _, kind := range scalable {
		if k == kind {
			return true
		}
	}

	return false
}

// @see k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta{}
type ObjectMeta struct {
	Name string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`
	UID types.UID `json:"uid,omitempty"`
}
func NewObjectMeta(k8SObjectMeta metav1.ObjectMeta) ObjectMeta {
	return ObjectMeta{
		Name:              k8SObjectMeta.Name,
		Namespace:         k8SObjectMeta.Namespace,
		Labels:            k8SObjectMeta.Labels,
		CreationTimestamp: k8SObjectMeta.CreationTimestamp,
		Annotations:       k8SObjectMeta.Annotations,
		UID:               k8SObjectMeta.UID,
	}
}

type TypeMeta struct {
	Kind ResourceKind `json:"kind,omitempty"`

	// Scalable represents whether or not an object is scalable.
	Scalable bool `json:"scalable,omitempty"`
}
func NewTypeMeta(kind ResourceKind) TypeMeta {
	return TypeMeta{
		Kind:     kind,
		Scalable: kind.Scalable(),
	}
}


type JsonResponse struct {
	Errno  int         `json:"errno"`  // -1,0
	Errmsg string      `json:"errmsg"` // "success" or "failed: xxx"
	Data   interface{} `json:"data"`   // struct{}
}
