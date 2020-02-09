package models

import "time"

type WebHookScope int64

const (
	WebHookScopeNamespace WebHookScope = iota
	TableNameWebHook                   = "web_hooks"
)

type webHookModel struct{}

type WebHook struct {
	ID        uint       `gorm:"column:id;primary_key;"`
	Name      string     `gorm:"column:name;size:128;not null;default:'';"`
	Scope     int        `gorm:"column:scope;not null;default:'0';"`
	ObjectId  uint       `gorm:"column:object_id"`
	Url       string     `gorm:"column:url;size:512;default:null;"`
	Secret    string     `gorm:"column:secret;size:512;default:null;"`
	Events    string     `gorm:"column:events;type:longtext;not null;"`
	Enabled   bool       `gorm:"column:enabled;not null;default:0;"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt *time.Time `gorm:"column:deleted_at;default:null;"`
}

func (WebHook) TableName() string {
	return TableNameWebHook
}
