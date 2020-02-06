package models

import (
	"fmt"
	"time"
)

const (
	GlobalAPIKey    APIKeyType = 0
	TableNameApiKey            = "api_keys"
)

type APIKeyType int32

type APIKey struct {
	ID         uint   `gorm:"column:id;primary_key;"`
	Name       string `gorm:"column:name;size:128;not null;index:api_key_name;default:'';"`
	Token      string `gorm:"column:token;type:longtext;not null;"`
	Type       uint   `gorm:"column:type;type:int;size:11;not null;default:0;"` // 0：全局 1：命名空间 2：项目
	ResourceId uint64 `gorm:"column:resource_id;type:bigint;size:20;default:null;"`
	//Group       *Group     `gorm:"null;rel(fk);on_delete(set_null)" json:"group,omitempty"`
	Description string    `gorm:"column:description;size:512;default:null;"`
	UserID      uint      `gorm:"column:user_id;"`
	GroupID     uint      `gorm:"column:group_id;"`
	ExpireIn    uint64    `gorm:"column:expire_in;type:bigint;not null;default:0;"` // 过期时间，单位：秒
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time `gorm:"column:deleted_at;default:null;"`

	User  User  `gorm:"foreignkey:UserID;association_foreignkey:ID;"`
	Group Group `gorm:"foreignkey:GroupID;association_foreignkey:ID;"`
	//Group Group `gorm:"foreignkey:GroupId;AssociationForeignKey:Id"`
	//User        User     `gorm:"foreignkey:UserId;AssociationForeignKey:Id"`
}

func (APIKey) TableName() string {
	return TableNameApiKey
}

func (k *APIKey) String() string {
	return fmt.Sprintf("[APIKey %d] %s", k.ID, k.Name)
}

/*func (k *APIKey) TableName() string {
	return TableNameApiKey
}*/

type apiKeyModel struct{}
