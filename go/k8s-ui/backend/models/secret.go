package models

import "time"

const (
	TableNameSecret = "secrets"
)

type secretModel struct{}

type Secret struct {
	ID          uint      `gorm:"column:id;primary_key;"`
	Name        string    `gorm:"column:name;size:128;not null;unique;default:'';"`
	MetaData    string    `gorm:"column:meta_data;type:longtext;not null;"`
	AppId       uint      `gorm:"column:app_id;size:20;not null;"`
	Description string    `gorm:"column:description;size:512;default:null;"`
	OrderId     uint      `gorm:"column:order_id;size:20;"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`

	//Id          int64  `gorm:"auto" json:"id,omitempty"`
	//Name        string `gorm:"unique;index;size(128)" json:"name,omitempty"`
	//MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	//App         *App   `gorm:"index;rel(fk)" json:"app,omitempty"`
	//Description string `gorm:"null;size(512)" json:"description,omitempty"`
	//OrderId     int64  `gorm:"index;default(0)" json:"order"`
	//
	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//User       string     `gorm:"size(128)" json:"user,omitempty"`
	//Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`
	//
	//AppId int64 `gorm:"-" json:"appId,omitempty"`
}

func (*Secret) TableName() string {
	return TableNameSecret
}
