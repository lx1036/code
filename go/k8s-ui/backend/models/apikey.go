package models

import (
	"fmt"
	"time"
)

const (
	GlobalAPIKey    APIKeyType = 0
	TableNameApiKey            = "api_key"
)

type APIKeyType int32

type APIKey struct {
	Id    int64  `gorm:"auto" json:"id,omitempty"`
	Name  string `gorm:"index;size(128)" json:"name,omitempty"`
	Token string `gorm:"type(text)" json:"token,omitempty"`
	// 0：全局 1：命名空间 2：项目
	Type       APIKeyType `gorm:"type(integer)" json:"type"`
	ResourceId int64      `gorm:"null;type(bigint)" json:"resourceId,omitempty"`
	// TODO beego 默认删除规则为级联删除，可选项 do_nothing on_delete
	Group       *Group     `gorm:"null;rel(fk);on_delete(set_null)" json:"group,omitempty"`
	Description string     `gorm:"null;size(512)" json:"description,omitempty"`
	User        string     `gorm:"size(128)" json:"user,omitempty"`
	ExpireIn    int64      `gorm:"type(bigint)" json:"expireIn"`            // 过期时间，单位：秒
	Deleted     bool       `gorm:"default(false)" json:"deleted,omitempty"` // 是否生效
	CreateTime  *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (k *APIKey) String() string {
	return fmt.Sprintf("[APIKey %d] %s", k.Id, k.Name)
}
func (k *APIKey) TableName() string {
	return TableNameApiKey
}

type apiKeyModel struct{}
