package models

import (
	"k8s.io/api/core/v1"
	"time"
)

type Namespace struct {
	Id            int64             `orm:"auto" json:"id,omitempty"`
	Name          string            `orm:"index;unique;size(128)" json:"name,omitempty"`
	KubeNamespace string            `orm:"index;size(128)" json:"kubeNamespace,omitempty"`
	MetaData      string            `orm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj   NamespaceMetaData `orm:"-" json:"-"`
	CreateTime    *time.Time        `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime    *time.Time        `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User          string            `orm:"size(128)" json:"user,omitempty"`
	Deleted       bool              `orm:"default(false)" json:"deleted,omitempty"`

	// 用于权限的关联查询
	NamespaceUsers []*NamespaceUser `orm:"reverse(many)" json:"-"`
}

type NamespaceMetaData struct {
	// key is cluster name, if the key not exist on clusterMeta
	// means this namespace could't use the cluster
	ClusterMetas map[string]ClusterMeta `json:"clusterMeta,omitempty"`
	// current namespace env, will overwrite cluster's Env
	Env []v1.EnvVar `json:"env,omitempty"`
	// current namespace image pull secrets, will overwrite cluster's ImagePullSecrets
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets"`
	// current namespace service annotation, will overwrite cluster service's Annotation
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// current namespace ingress annotation, will overwrite cluster ingress's Annotation
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
}

type ClusterMeta struct {
	ResourcesLimit ResourcesLimit `json:"resourcesLimit"`
}

type ResourcesLimit struct {
	// unit core
	Cpu int64 `json:"cpu,omitempty"`
	// unit G
	Memory int64 `json:"memory,omitempty"`
}


type namespaceModel struct{}

func (model *namespaceModel) GetByName(name string) (namespace *Namespace, err error) {
	namespace = &Namespace{Name: name}
	if err = Ormer().Read(namespace, "name"); err == nil {
		return namespace, nil
	}

	return nil, err
}

