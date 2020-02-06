package models

import "time"

const (
	TableNameNotification = "notifications"
)

type NotificationType string
type NotificationLevel int

type Notification struct {
	Id      int64            `gorm:"column:id;primary_key;"`
	Type    NotificationType `gorm:"column:type;size:128;not null;default:'';"`
	Title   string           `gorm:"column:title;size:2000;not null;default:'';"`
	Message string           `gorm:"column:message;type:longtext;not null;"`
	//FromUser    *User             `gorm:"index;rel(fk)" json:"from,omitempty"`
	FromUserId  uint              `gorm:"column:from_user_id"`
	Level       NotificationLevel `gorm:"column:level;size:11;not null;default:'0';"`
	IsPublished bool              `gorm:"column:is_published;not null;default:'0';"`
	CreatedAt   time.Time         `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   time.Time         `gorm:"column:deleted_at;default:null;"`
	//CreateTime  *time.Time        `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	//UpdateTime  *time.Time        `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (Notification) TableName() string {
	return TableNameNotification
}

type notificationModel struct{}
