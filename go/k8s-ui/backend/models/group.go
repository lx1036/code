package models

import "time"

type GroupType int

const (
	AppGroupType GroupType = iota
	NamespaceGroupType
	TableNameGroup = "group"
)

type Group struct {
	Id      int64     `orm:"pk;auto" json:"id,omitempty"`
	Name    string    `orm:"index;size(200)" json:"name,omitempty"`
	Comment string    `orm:"type(text)" json:"comment,omitempty"`
	Type    GroupType `orm:"type(integer)" json:"type"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	// 用于权限的关联查询
	Permissions    []*Permission    `orm:"rel(m2m);rel_table(group_permissions)" json:"permissions,omitempty"`
	AppUsers       []*AppUser       `orm:"reverse(many)" json:"appUsers,omitempty"`
	NamespaceUsers []*NamespaceUser `orm:"reverse(many)" json:"namespaceUsers,omitempty"`
}

func (*Group) TableName() string {
	return TableNameGroup
}

type groupModel struct{}
