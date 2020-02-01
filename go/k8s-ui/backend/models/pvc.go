package models

import "time"

const (
	TableNamePersistentVolumeClaim = "persistent_volume_claim"
)

type PersistentVolumeClaim struct {
	Id          int64  `gorm:"auto" json:"id,omitempty"`
	Name        string `gorm:"unique;index;size(128)" json:"name,omitempty"`
	MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	App         *App   `gorm:"index;rel(fk)" json:"app,omitempty"`
	Description string `gorm:"null;size(512)" json:"description,omitempty"`
	OrderId     int64  `gorm:"index;default(0)" json:"order"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`

	AppId int64 `gorm:"-" json:"appId,omitempty"`
}

func (*PersistentVolumeClaim) TableName() string {
	return TableNamePersistentVolumeClaim
}

type persistentVolumeClaimModel struct{}
