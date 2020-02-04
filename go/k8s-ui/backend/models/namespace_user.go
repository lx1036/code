package models

import "time"

const (
	TableNameNamespaceUser = "namespace_users"
)

type NamespaceUser struct {
	Id          uint      `gorm:"column:id;primary_key;"`
	NamespaceID uint      `gorm:"column:namespace_id;"`
	UserID      uint      `gorm:"column:user_id;"`
	GroupID     uint      `gorm:"column:group_id;"`
	CreatedAt   time.Time `gorm:"column:created_at;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`

	//User        User      `gorm:"foreignkey:UserID;association_foreignkey:ID;"`
	//Namespace       Group     `gorm:"foreignkey:GroupID;association_foreignkey:ID;"`

	//Namespace *Namespace `gorm:"index;rel(fk);column(namespace_id)" json:"namespace,omitempty"`
	//User      *User      `gorm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
	//Group     *Group     `gorm:"index;rel(fk)" json:"group,omitempty"`

	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	//Groups     []*Group `gorm:"-" json:"groups,omitempty"`
	//GroupsName string   `gorm:"-" json:"groupsName,omitempty"`
}

func (NamespaceUser) TableName() string {
	return TableNameNamespaceUser
}

type namespaceUserModel struct{}

func (namespaceUser *namespaceUserModel) GetAllPermissions(permissionId int64, userId int64) (permissions []Permission, err error) {
	qs := Ormer().QueryTable(&Permission{}).Filter("Groups__Group__Type__exact", NamespaceGroupType).
		Filter("Groups__Group__NamespaceUsers__Namespace__Id__exact", permissionId).
		Filter("Groups__Group__NamespaceUsers__User__Id__exact", userId)
	permissions = []Permission{}
	if _, err = qs.All(permissions); err != nil {
		return
	}
	return
}
