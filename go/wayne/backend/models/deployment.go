package models

import (
	"time"
	kapi "k8s.io/api/core/v1"
)


type DeploymentMetaData struct {
	Replicas  map[string]int32  `json:"replicas"`
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type Deployment struct {
	Id   int64  `orm:"auto" json:"id,omitempty"`
	Name string `orm:"unique;index;size(128)" json:"name,omitempty"`
	MetaData    string             `orm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj DeploymentMetaData `orm:"-" json:"-"`
	App         *App               `orm:"index;rel(fk)" json:"app,omitempty"`
	Description string             `orm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64              `orm:"index;default(0)" json:"order"`
	
	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `orm:"default(false)" json:"deleted,omitempty"`
	
	AppId int64 `orm:"-" json:"appId,omitempty"`
}

type deploymentModel struct{}

func (*deploymentModel) GetByName(name string) (deployment *Deployment, err error)  {
	deployment = &Deployment{
		Id:          0,
		Name:        name,
		MetaData:    "",
		MetaDataObj: DeploymentMetaData{},
		App:         nil,
		Description: "",
		OrderId:     0,
		CreateTime:  nil,
		UpdateTime:  nil,
		User:        "",
		Deleted:     false,
		AppId:       0,
	}
	if err := Ormer().Read(deployment, "name"); err == nil {
		deployment.Id = deployment.App.Id
		return deployment, nil
	}

	return nil, err
}
