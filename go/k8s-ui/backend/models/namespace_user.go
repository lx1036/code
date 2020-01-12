package models

import "time"

const (
	TableNameNamespaceUser = "namespace_user"
)

type NamespaceUser struct {
	Id        int64      `orm:"auto" json:"id,omitempty"`
	Namespace *Namespace `orm:"index;rel(fk);column(namespace_id)" json:"namespace,omitempty"`
	User      *User      `orm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
	Group     *Group     `orm:"index;rel(fk)" json:"group,omitempty"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Groups     []*Group `orm:"-" json:"groups,omitempty"`
	GroupsName string   `orm:"-" json:"groupsName,omitempty"`
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
