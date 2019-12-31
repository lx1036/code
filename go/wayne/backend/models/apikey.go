package models

import (
	"fmt"
	"time"
)

const (
	GlobalAPIKey      APIKeyType = 0
)

type APIKeyType int32

type APIKey struct {
	Id    int64  `orm:"auto" json:"id,omitempty"`
	Name  string `orm:"index;size(128)" json:"name,omitempty"`
	Token string `orm:"type(text)" json:"token,omitempty"`
	// 0：全局 1：命名空间 2：项目
	Type       APIKeyType `orm:"type(integer)" json:"type"`
	ResourceId int64      `orm:"null;type(bigint)" json:"resourceId,omitempty"`
	// TODO beego 默认删除规则为级联删除，可选项 do_nothing on_delete
	Group       *Group     `orm:"null;rel(fk);on_delete(set_null)" json:"group,omitempty"`
	Description string     `orm:"null;size(512)" json:"description,omitempty"`
	User        string     `orm:"size(128)" json:"user,omitempty"`
	ExpireIn    int64      `orm:"type(bigint)" json:"expireIn"`            // 过期时间，单位：秒
	Deleted     bool       `orm:"default(false)" json:"deleted,omitempty"` // 是否生效
	CreateTime  *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (k *APIKey) String() string {
	return fmt.Sprintf("[APIKey %d] %s", k.Id, k.Name)
}
