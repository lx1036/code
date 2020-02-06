package models

import (
	kapi "k8s.io/api/core/v1"
	"time"
)

const (
	TableNameCronjob = "cronjobs"
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
	ID          uint      `gorm:"column:id;primary_key;"`
	Name        string    `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string    `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint      `gorm:"column:app_id;size:20;not null;"`
	Description string    `gorm:"column:description;size:512;default:null;"`
	OrderId     uint      `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`

	// 存储模版可上线机房，已挂起的机房
	/*
		{
		  "replicas": {
		    "K8S": 1
		  },
		}
	*/
	//App         *App            `gorm:"index;rel(fk)" json:"app,omitempty"`
	//Description string          `gorm:"null;size(512)" json:"description,omitempty"`
	//OrderId     int64           `gorm:"index;default(0)" json:"order"`
	//
	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User       string     `gorm:"size(128)" json:"user,omitempty"`
	//Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`
	//
	//AppId int64 `gorm:"-" json:"appId,omitempty"`
}

func (Cronjob) TableName() string {
	return TableNameCronjob
}

func (*cronjobModel) GetById(id int64) (v *Cronjob, err error) {
	v = &Cronjob{ID: uint(id)}
	if err = Ormer().Read(v); err == nil {
		//v.AppId = uint(v.App.ID)
		return v, nil
	}
	return nil, err
}
