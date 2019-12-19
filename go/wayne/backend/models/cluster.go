package models

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type ClusterStatus int32


type Cluster struct {
	Id          int64  `orm:"auto" json:"id,omitempty"`
	Name        string `orm:"unique;index;size(128)" json:"name,omitempty"`
	DisplayName string `orm:"size(512);column(displayname);null" json:"displayname,omitempty"`
	MetaData    string     `orm:"null;type(text)" json:"metaData,omitempty"`
	Master      string     `orm:"size(128)" json:"master,omitempty"` // apiserver地址，示例： https://10.172.189.140
	KubeConfig  string     `orm:"null;type(text)" json:"kubeConfig,omitempty"`
	Description string     `orm:"null;size(512)" json:"description,omitempty"`
	CreateTime  *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User        string     `orm:"size(128)" json:"user,omitempty"`
	Deleted     bool       `orm:"default(false)" json:"deleted,omitempty"`
	// the cluster status
	Status ClusterStatus `orm:"default(0)" json:"status"`
	
	MetaDataObj ClusterMetaData `orm:"-" json:"-"`
}

type ClusterMetaData struct {
	// robin plugin
	Robin *ClusterRobinMetaData `json:"robin"`
	// kubetool log source
	LogSource string `json:"logSource"`
	// rbd默认配置，创建或修改RBD类型的PV时会使用此配置填充
	RBD *v1.RBDVolumeSource `json:"rbd"`
	// cephfs默认配置，创建或修改cephfs类型的PV时会使用此配置填充
	CephFS *v1.CephFSVolumeSource `json:"cephfs"`
	// 默认添加环境变量，会在发布资源时在每个Container添加此环境变量, will be overwrite by namespace's Env
	Env []v1.EnvVar
	// current cluster image pull secrets, will be overwrite by namespace's ImagePullSecrets
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets"`
	// 默认添加service注解，会在发布资源时在每个service添加此Annotations, will be overwrite by namespace's Annotations
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// 默认添加ingress注解，会在发布资源时在每个ingress添加此Annotations, will be overwrite by namespace's Annotations
	IngressAnnotations map[string]string `json:"ingressAnnotations,omitempty"`
}

type ClusterRobinMetaData struct {
	Token          string `json:"token"`
	Url            string `json:"url"`
	SftpPort       int    `json:"sftpPort"`
	PasswordDesKey string `json:"passwordDesKey"`
}
