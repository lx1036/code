package models

import (
	"k8s.io/api/core/v1"
	"time"
)

const (
	TableNameNamespace = "namespaces"
)

type Namespace struct {
	ID            uint      `gorm:"column:id;primary_key;"`
	Name          string    `gorm:"column:name;size:128;not null;unique;default:'';"`
	KubeNamespace string    `gorm:"column:kube_namespace;size:128;not null;default:'';"`
	MetaData      string    `gorm:"column:meta_data;type:longtext;not null;"`
	CreatedAt     time.Time `gorm:"column:created_at;"`
	UpdatedAt     time.Time `gorm:"column:updated_at;"`
	DeletedAt     time.Time `gorm:"column:deleted_at;default:null;"`

	Users []*User `gorm:"many2many:namespace_users;"`
	//MetaDataObj   NamespaceMetaData `gorm:"-" json:"-"`
	//CreateTime    *time.Time        `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime    *time.Time        `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User          string            `gorm:"size(128)" json:"user,omitempty"`
	//Deleted       bool              `gorm:"default(false)" json:"deleted,omitempty"`
	//
	//// 用于权限的关联查询
	//NamespaceUsers []*NamespaceUser `gorm:"reverse(many)" json:"-"`
}

func (Namespace) TableName() string {
	return TableNameNamespace
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

func (model *namespaceModel) GetAll(deleted bool) ([]*Namespace, error) {
	var namespaces []*Namespace
	_, err := Ormer().QueryTable(new(Namespace)).Filter("Deleted", deleted).OrderBy("Name").All(&namespaces)
	if err != nil {
		return nil, err
	}

	return namespaces, nil
}

func (model *namespaceModel) GetById(id int64) (v *Namespace, err error) {
	v = &Namespace{ID: uint(id)}

	if err = Ormer().Read(v); err == nil {
		return v, nil
	}
	return nil, err
}
