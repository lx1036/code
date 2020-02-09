package models

import "time"

type AppUser struct {
	Id        uint      `gorm:"column:id;primary_key;"`
	AppID     uint      `gorm:"column:app_id;"`
	GroupID   uint      `gorm:"column:group_id;"`
	UserID    uint      `gorm:"column:user_id;"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`

	//App   *App   `gorm:"index;rel(fk);column(app_id)" json:"app,omitempty"`
	//User  *User  `gorm:"index;rel(fk);column(user_id)" json:"user,omitempty"`
	//Group *Group `gorm:"index;rel(fk)" json:"group,omitempty"`
	//
	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime *time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	//
	//Groups     []*Group `gorm:"-" json:"groups,omitempty"`
	//GroupsName string   `gorm:"-" json:"groupsName,omitempty"`
}

type appUserModel struct{}
