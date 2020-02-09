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
	ID        uint       `gorm:"column:id;primary_key;" json:"id"`
	Name      string     `gorm:"column:name;size:200;not null;unique;default:'';" json:"name"`
	Password  string     `gorm:"column:password;size:255;not null;default:'';" json:"password"`
	Salt      string     `gorm:"column:salt;size:32;not null;default:'';" json:"salt"`
	Email     string     `gorm:"column:email;size:200;not null;default:'';" json:"email"`
	Display   string     `gorm:"column:display;size:200;not null;default:'';" json:"display"`
	Comment   string     `gorm:"column:comment;type:longtext;not null;" json:"comment"`
	Type      uint       `gorm:"column:type;size:11;not null;default:0;" json:"type"`
	Admin     bool       `gorm:"column:admin;not null;default:0;" json:"admin"`
	LastLogin time.Time  `gorm:"column:last_login;not null;" json:"last_login"`
	LastIp    string     `gorm:"column:last_ip;size:200;not null;default:'';" json:"last_ip"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:current_timestamp;" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;default:null;" json:"deleted_at"`

	ApiKeys       []APIKey       `gorm:"foreignkey:UserID;association_foreignkey:ID;" json:"api_keys,omitempty"`
	Namespaces    []Namespace    `gorm:"many2many:namespace_users;" json:"namespaces,omitempty"`
	Notifications []Notification `gorm:"column:notifications;foreignkey:FromUserId;" json:"notifications,omitempty"`
}

func (User) TableName() string {
	return TableNameUser
}

type UserStatistics struct {
	Total int64 `json:"total,omitempty"`
}

type userModel struct{}

func GetUserByName(name string) (user User, err error) {
	//user = &User{Name: name}
	//if err = Ormer().Read(user, "Name"); err != nil {
	//	return nil, err
	//}
	//var user models.User

	err = lorm.DB.Where("name=?", name).First(&user).Error
	if err != nil {
		return User{}, err
	}

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
