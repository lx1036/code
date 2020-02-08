package models

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	ClusterStatusNormal ClusterStatus = 0
	TableNameCluster                  = "clusters"
)

type ClusterStatus int32

type Cluster struct {
	ID          uint       `gorm:"column:id;primary_key;"`
	Name        string     `gorm:"column:name;size:128;not null;unique;default:'';"`
	DisplayName string     `gorm:"column:display_name;size:512;default:null;"`
	MetaData    string     `gorm:"column:meta_data;type:longtext;default:null;"`
	Master      string     `gorm:"column:master;size:128;not null;default:'';"` // apiserver地址，示例： https://10.172.189.140
	KubeConfig  string     `gorm:"column:kube_config;type:longtext;default:null;"`
	Description string     `gorm:"column:description;size:512;default:null;"`
	Status      int        `gorm:"column:status;size:11;not null;default:0;"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`

	//User        string     `gorm:"size(128)" json:"user,omitempty"`

	//Status ClusterStatus `gorm:"default(0)" json:"status"`

	//MetaDataObj ClusterMetaData `gorm:"-" json:"-"`
}

func (Cluster) TableName() string {
	return TableNameCluster
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

type clusterModel struct{}

func (model *clusterModel) GetParsedMetaDataByName(name string) (v *Cluster, err error) {
	v = &Cluster{Name: name}
	if err = Ormer().Read(v, "Name"); err == nil {
		if v.MetaData != "" {
			//err := json.Unmarshal(hack.Slice(v.MetaData), &v.MetaDataObj)
			//if err != nil {
			//
			//}
		}

		return v, nil
	}

	return nil, err
}

func (model *clusterModel) GetAllNormal() ([]Cluster, error) {
	clusters := []Cluster{}
	_, err := Ormer().
		QueryTable(new(Cluster)).
		Filter("Status", ClusterStatusNormal).
		Filter("Deleted", false).
		All(&clusters)

	if err != nil {
		return nil, err
	}

	return clusters, nil
}
