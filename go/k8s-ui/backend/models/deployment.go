package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameDeployment = "deployment"
)

type DeploymentMetaData struct {
	Replicas  map[string]int32  `json:"replicas"`
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type Deployment struct {
	Id          int64              `gorm:"auto" json:"id,omitempty"`
	Name        string             `gorm:"unique;index;size(128)" json:"name,omitempty"`
	MetaData    string             `gorm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj DeploymentMetaData `gorm:"-" json:"-"`
	App         *App               `gorm:"index;rel(fk)" json:"app,omitempty"`
	Description string             `gorm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64              `gorm:"index;default(0)" json:"order"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`

	AppId int64 `gorm:"-" json:"appId,omitempty"`
}

type deploymentModel struct{}

func (*deploymentModel) GetByName(name string) (deployment *Deployment, err error) {
	deployment = &Deployment{Name: name}
	if err := Ormer().Read(deployment, "name"); err == nil {
		deployment.Id = int64(deployment.App.ID)
		return deployment, nil
	}

	return nil, err
}

func (model *deploymentModel) GetById(id int64) (deployment *Deployment, err error) {
	deployment = &Deployment{Id: id}

	if err = Ormer().Read(deployment); err == nil {
		deployment.AppId = int64(deployment.App.ID)
		return deployment, nil
	}

	return nil, err
}

func (model *deploymentModel) UpdateById(deployment *Deployment) (err error) {
	v := Deployment{Id: deployment.Id}
	// ascertain id exists in the database
	if err = Ormer().Read(&v); err == nil {
		deployment.App = &App{ID: uint(deployment.AppId)}
		deployment.UpdateTime = nil
		_, err = Ormer().Update(deployment)
		return err
	}

	return nil
}
