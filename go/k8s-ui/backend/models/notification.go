package models

import "time"

const (
	TableNameNotification = "notification"
)

type NotificationType string
type NotificationLevel int

type Notification struct {
	Id          int64             `gorm:"auto" json:"id,omitempty"`
	Type        NotificationType  `gorm:"index;size(128)" json:"type,omitempty"`
	Title       string            `gorm:"size(2000)" json:"title,omitempty"`
	Message     string            `gorm:"type(text)" json:"message,omitempty"`
	FromUser    *User             `gorm:"index;rel(fk)" json:"from,omitempty"`
	Level       NotificationLevel `gorm:"default(0)" json:"level,omitempty"`
	IsPublished bool              `gorm:"default(false)" json:"is_published"`
	CreateTime  *time.Time        `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time        `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (*Notification) TableName() string {
	return TableNameNotification
}

type notificationModel struct{}
