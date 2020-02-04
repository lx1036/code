package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameCronjob = "cronjob"
)

type cronjobModel struct{}

type CronjobMetaData struct {
	Replicas map[string]int32 `json:"replicas"`
	Suspends map[string]bool  `json:"suspends"`
	Affinity *kapi.Affinity   `json:"affinity,omitempty"`
	// 是否允许用户使用特权模式，默认不允许,key 为容器名称
	Privileged map[string]*bool `json:"privileged"`
}

type Cronjob struct {
	Id   int64  `gorm:"auto" json:"id,omitempty"`
	Name string `gorm:"unique;size(128)" json:"name,omitempty"`
	// 存储模版可上线机房，已挂起的机房
	/*
		{
		  "replicas": {
		    "K8S": 1
		  },
		}
	*/
	MetaData    string          `gorm:"type(text)" json:"metaData,omitempty"`
	MetaDataObj CronjobMetaData `gorm:"-" json:"-"`
	App         *App            `gorm:"index;rel(fk)" json:"app,omitempty"`
	Description string          `gorm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64           `gorm:"index;default(0)" json:"order"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`

	AppId int64 `gorm:"-" json:"appId,omitempty"`
}

func (*cronjobModel) GetById(id int64) (v *Cronjob, err error) {
	v = &Cronjob{Id: id}
	if err = Ormer().Read(v); err == nil {
		v.AppId = int64(v.App.ID)
		return v, nil
	}
	return nil, err
}
