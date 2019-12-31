package models

import "time"

type AppUser struct {
	Id    int64  `orm:"auto" json:"id,omitempty"`
	App   *App   `orm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	User  *User  `orm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
	Group *Group `orm:"index;rel(fk)" json:"group,omitempty"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Groups     []*Group `orm:"-" json:"groups,omitempty"`
	GroupsName string   `orm:"-" json:"groupsName,omitempty"`
}
