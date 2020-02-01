package models

import "time"

type AppUser struct {
	Id    int64  `gorm:"auto" json:"id,omitempty"`
	App   *App   `gorm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	User  *User  `gorm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
	Group *Group `gorm:"index;rel(fk)" json:"group,omitempty"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Groups     []*Group `gorm:"-" json:"groups,omitempty"`
	GroupsName string   `gorm:"-" json:"groupsName,omitempty"`
}

type appUserModel struct{}
