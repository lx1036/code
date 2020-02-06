package models

import "time"

const (
	TableNameNotificationLog = "notification_logs"
)

type NotificationLog struct {
	Id             uint      `gorm:"column:id;primary_key;"`
	UserId         uint      `gorm:"column:user_id;"`
	NotificationId uint      `gorm:"column:notification_id;"`
	IsReaded       bool      `gorm:"column:is_readed;not null;default:0;"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt      time.Time `gorm:"column:deleted_at;default:null;"`
}

func (NotificationLog) TableName() string {
	return TableNameNotificationLog
}

type notificationLogModel struct{}
