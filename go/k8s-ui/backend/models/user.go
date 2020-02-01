package models

import (
	"github.com/astaxie/beego/orm"
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"time"
)

type UserType int

type User struct {
	Id         int64      `gorm:"pk;auto" json:"id,omitempty"`
	Name       string     `gorm:"index;unique;size(200)" json:"name,omitempty"`
	Password   string     `gorm:"size(255)" json:"-"`
	Salt       string     `gorm:"size(32)" json:"-"`
	Email      string     `gorm:"size(200)" json:"email,omitempty"`
	Display    string     `gorm:"size(200)" json:"display,omitempty"`
	Comment    string     `gorm:"type(text)" json:"comment,omitempty"`
	Type       UserType   `gorm:"type(integer)" json:"type"`
	Admin      bool       `gorm:"default(False)" json:"admin"`
	LastLogin  *time.Time `gorm:"auto_now_add;type(datetime)" json:"lastLogin,omitempty"`
	LastIp     string     `gorm:"size(200)" json:"lastIp,omitempty"`
	Deleted    bool       `gorm:"default(false)" json:"deleted,omitempty"`
	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Namespaces []*Namespace `gorm:"-" json:"namespaces,omitempty"`
}

type UserStatistics struct {
	Total int64 `json:"total,omitempty"`
}

type userModel struct{}

func (model *userModel) GetUserByName(name string) (user *User, err error) {
	user = &User{Name: name}
	if err = Ormer().Read(user, "Name"); err != nil {
		return nil, err
	}

	return user, nil
}

func (model *userModel) GetUserDetail(name string) (user *User, err error) {
	lorm.DB.Where("name = ?", name).First(&user)


	/*user = &User{Name: name}
	err = Ormer().Read(user, "Name")
	if err != nil {
		return nil, err
	}*/
	if user.Admin {
		namespaces, err := NamespaceModel.GetAll(false)
		if err != nil {
			return nil, err
		}
		user.Namespaces = namespaces
	} else {
		var namespaceUsers []NamespaceUser
		condNS := (orm.NewCondition()).And("User__Id__exact", user.Id)
		_, err = Ormer().QueryTable(TableNameNamespaceUser).
			SetCond(condNS).
			RelatedSel("Namespace").
			GroupBy("Namespace").
			OrderBy("Namespace__Name").
			All(&namespaceUsers)
		if err != nil {
			return nil, err
		}
		for _, namespaceUser := range namespaceUsers {
			user.Namespaces = append(user.Namespaces, namespaceUser.Namespace)
		}
	}

	return user, nil
}
