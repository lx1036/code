package models

import "time"

const (
	TableNameNotificationLog = "notification_log"
)

type NotificationLog struct {
	Id           int64         `gorm:"auto" json:"id,omitempty"`
	UserId       int64         `gorm:"default(0)" json:"user_id,omitempty"`
	CreateTime   *time.Time    `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	IsReaded     bool          `gorm:"default(false)" json:"is_readed"`
	Notification *Notification `gorm:"index;column(notification_id);rel(fk)" json:"notification"`
}

func (*NotificationLog) TableName() string {
	return TableNameNotificationLog
}

type notificationLogModel struct{}
