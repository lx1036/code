package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameDeployment = "deployments"
)

type DeploymentMetaData struct {
	Replicas  map[string]int32  `json:"replicas"`
	Resources map[string]string `json:"resources,omitempty"`
	Affinity  *kapi.Affinity    `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type Deployment struct {
	ID          uint       `gorm:"column:id;primary_key;"`
	Name        string     `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string     `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint       `gorm:"column:app_id;size:20;not null;"`
	Description string     `gorm:"column:description;size:512;default:null;"`
	OrderId     uint       `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`
}

func (Deployment) TableName() string {
	return TableNameDeployment
}

type deploymentModel struct{}

func (*deploymentModel) GetByName(name string) (deployment *Deployment, err error) {
	deployment = &Deployment{Name: name}
	if err := Ormer().Read(deployment, "name"); err == nil {
		//deployment.ID = int64(deployment.App.ID)
		return deployment, nil
	}

	return nil, err
}

func (model *deploymentModel) GetById(id int64) (deployment *Deployment, err error) {
	//deployment = &Deployment{Id: id}
	//
	//if err = Ormer().Read(deployment); err == nil {
	//	deployment.AppId = int64(deployment.App.ID)
	//	return deployment, nil
	//}

	return nil, err
}

func (model *deploymentModel) UpdateById(deployment *Deployment) (err error) {
	//v := Deployment{Id: deployment.Id}
	//// ascertain id exists in the database
	//if err = Ormer().Read(&v); err == nil {
	//	deployment.App = &App{ID: uint(deployment.AppId)}
	//	deployment.UpdateTime = nil
	//	_, err = Ormer().Update(deployment)
	//	return err
	//}

	return nil
}
