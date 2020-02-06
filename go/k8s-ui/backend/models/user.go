package models

import (
	"k8s-lx1036/k8s-ui/backend/database/lorm"
	"time"
)

type UserType int

const (
	TableNameUser = "users"
)

type User struct {
	ID        uint      `gorm:"column:id;primary_key;"`
	Name      string    `gorm:"column:name;size:200;not null;unique;default:'';"`
	Password  string    `gorm:"column:password;size:255;not null;default:'';"`
	Salt      string    `gorm:"column:salt;size:32;not null;default:'';"`
	Email     string    `gorm:"column:email;size:200;not null;default:'';"`
	Display   string    `gorm:"column:display;size:200;not null;default:'';"`
	Comment   string    `gorm:"column:comment;type:longtext;not null;"`
	Type      uint      `gorm:"column:type;size:11;not null;default:0;"`
	Admin     bool      `gorm:"column:admin;not null;default:0;"`
	LastLogin time.Time `gorm:"column:last_login;not null;"`
	LastIp    string    `gorm:"column:last_ip;size:200;not null;default:'';"`
	CreatedAt time.Time `gorm:"column:created_at;"`
	UpdatedAt time.Time `gorm:"column:updated_at;"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:null;"`

	ApiKeys []APIKey `gorm:"foreignkey:UserID;association_foreignkey:ID;"`
	//Namespace  Namespace      `gorm:"foreignkey:UserID;association_foreignkey:ID;"`
	Namespaces []*Namespace `gorm:"many2many:namespace_users;"`
}

func (User) TableName() string {
	return TableNameUser
}

type UserStatistics struct {
	Total int64 `json:"total,omitempty"`
}

type userModel struct{}

func (model *userModel) GetUserByName(name string) (user *User, err error) {
	user = &User{Name: name}
	//if err = Ormer().Read(user, "Name"); err != nil {
	//	return nil, err
	//}

	return user, nil
}

func (model *userModel) GetUserDetail(name string) (user *User, err error) {
	lorm.DB.Where("name = ?", name).First(&user)

	//if user.Admin {
	//	namespaces, err := NamespaceModel.GetAll(false)
	//	if err != nil {
	//		return nil, err
	//	}
	//	user.Namespaces = namespaces
	//} else {
	//	var namespaceUsers []NamespaceUser
	//	condNS := (orm.NewCondition()).And("User__Id__exact", user.Id)
	//	_, err = Ormer().QueryTable(TableNameNamespaceUser).
	//		SetCond(condNS).
	//		RelatedSel("Namespace").
	//		GroupBy("Namespace").
	//		OrderBy("Namespace__Name").
	//		All(&namespaceUsers)
	//	if err != nil {
	//		return nil, err
	//	}
	//	for _, namespaceUser := range namespaceUsers {
	//		user.Namespaces = append(user.Namespaces, namespaceUser.Namespace)
	//	}
	//}

	return user, nil
}
