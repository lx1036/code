package models

import "time"

const (
	TableNameNotification = "notification"
)

type NotificationType string
type NotificationLevel int

type Notification struct {
	Id          int64             `orm:"auto" json:"id,omitempty"`
	Type        NotificationType  `orm:"index;size(128)" json:"type,omitempty"`
	Title       string            `orm:"size(2000)" json:"title,omitempty"`
	Message     string            `orm:"type(text)" json:"message,omitempty"`
	FromUser    *User             `orm:"index;rel(fk)" json:"from,omitempty"`
	Level       NotificationLevel `orm:"default(0)" json:"level,omitempty"`
	IsPublished bool              `orm:"default(false)" json:"is_published"`
	CreateTime  *time.Time        `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  *time.Time        `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
}

func (*Notification) TableName() string {
	return TableNameNotification
}

type notificationModel struct{}
