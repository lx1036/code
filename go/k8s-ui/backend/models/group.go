package models

import "time"

type GroupType int

const (
	AppGroupType GroupType = iota
	NamespaceGroupType
	TableNameGroup = "groups"
)

type Group struct {
	ID        uint      `gorm:"column:id;primary_key;"`
	Name      string    `gorm:"column:name;size:200;not null;default:'';"`
	Comment   string    `gorm:"column:comment;type:longtext;not null;"`
	Type      uint      `gorm:"column:type;size:11;not null;default:0;"`
	CreatedAt time.Time `gorm:"column:created_at;"`
	UpdatedAt time.Time `gorm:"column:updated_at;"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:null;"`

	//ApiKeys []APIKey `gorm:"foreignkey:GroupId"`
	//ApiKeys []APIKey `gorm:"foreignkey:GroupID;AssociationForeignKey:ID"`
	ApiKeys []APIKey `gorm:"foreignkey:GroupID;association_foreignkey:ID;"`

	// 用于权限的关联查询
	//Permissions    []*Permission    `gorm:"rel(m2m);rel_table(group_permissions)" json:"permissions,omitempty"`
	//AppUsers       []*AppUser       `gorm:"reverse(many)" json:"appUsers,omitempty"`
	//NamespaceUsers []*NamespaceUser `gorm:"reverse(many)" json:"namespaceUsers,omitempty"`
}

func (Group) TableName() string {
	return TableNameGroup
}

type groupModel struct{}
