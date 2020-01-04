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
	Id       int64        `orm:"auto" json:"id"`
	Name     string       `orm:"index;size(128)" json:"name"`
	Scope    WebHookScope `json:"scope"`
	ObjectId int64        `json:"objectId"`

	Url    string `orm:"null;size(512)" json:"url"`
	Secret string `orm:"null;size(512)" json:"secret"`
	Events string `orm:"type(text)" json:"events"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime"`
	User       string     `orm:"size(128)" json:"user"`
	Enabled    bool       `orm:"default(false)" json:"enabled"`
}

func (*WebHook) TableName() string {
	return TableNameWebHook
}
