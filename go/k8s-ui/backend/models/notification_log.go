package models

import "time"

const (
	TableNameNotificationLog = "notification_log"
)

type NotificationLog struct {
	Id           int64         `orm:"auto" json:"id,omitempty"`
	UserId       int64         `orm:"default(0)" json:"user_id,omitempty"`
	CreateTime   *time.Time    `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	IsReaded     bool          `orm:"default(false)" json:"is_readed"`
	Notification *Notification `orm:"index;column(notification_id);rel(fk)" json:"notification"`
}

func (*NotificationLog) TableName() string {
	return TableNameNotificationLog
}

type notificationLogModel struct{}
