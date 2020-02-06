package models

import "time"

const (
	TableNameConfigMap = "config_maps"
)

type ConfigMap struct {
	ID          uint      `gorm:"column:id;primary_key;"`
	Name        string    `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string    `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint      `gorm:"column:app_id;size:20;not null;"`
	Description string    `gorm:"column:description;size:512;default:null;"`
	OrderId     uint      `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`
}

func (ConfigMap) TableName() string {
	return TableNameConfigMap
}

type configMapModel struct{}
