package models

import "time"

type GroupType int

const (
	AppGroupType GroupType = iota
	NamespaceGroupType
	TableNameGroup = "group"
)

type Group struct {
	Id      int64     `gorm:"pk;auto" json:"id,omitempty"`
	Name    string    `gorm:"index;size(200)" json:"name,omitempty"`
	Comment string    `gorm:"type(text)" json:"comment,omitempty"`
	Type    GroupType `gorm:"type(integer)" json:"type"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	// 用于权限的关联查询
	Permissions    []*Permission    `gorm:"rel(m2m);rel_table(group_permissions)" json:"permissions,omitempty"`
	AppUsers       []*AppUser       `gorm:"reverse(many)" json:"appUsers,omitempty"`
	NamespaceUsers []*NamespaceUser `gorm:"reverse(many)" json:"namespaceUsers,omitempty"`
}

func (*Group) TableName() string {
	return TableNameGroup
}

type groupModel struct{}
