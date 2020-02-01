package models

import "time"

type WebHookScope int64

const (
	WebHookScopeNamespace WebHookScope = iota
	WebHookScopeApp

	TableNameWebHook = "web_hook"
)

type webHookModel struct{}

type WebHook struct {
	Id       int64        `gorm:"auto" json:"id"`
	Name     string       `gorm:"index;size(128)" json:"name"`
	Scope    WebHookScope `json:"scope"`
	ObjectId int64        `json:"objectId"`

	Url    string `gorm:"null;size(512)" json:"url"`
	Secret string `gorm:"null;size(512)" json:"secret"`
	Events string `gorm:"type(text)" json:"events"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime"`
	User       string     `gorm:"size(128)" json:"user"`
	Enabled    bool       `gorm:"default(false)" json:"enabled"`
}

func (*WebHook) TableName() string {
	return TableNameWebHook
}
