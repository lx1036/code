package models

import (
	"time"
)

const (
	TableNameNotificationLog = "notification_logs"
)

type NotificationLog struct {
	ID             uint       `gorm:"column:id;primary_key;" json:"id"`
	UserId         uint       `gorm:"column:user_id;" json:"user_id"`
	NotificationId uint       `gorm:"column:notification_id;" json:"notification_id"`
	IsRead         bool       `gorm:"column:is_read;not null;default:0;" json:"is_read"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:current_timestamp;" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;default:null;" json:"deleted_at"`
}

func (NotificationLog) TableName() string {
	return TableNameNotificationLog
}

type notificationLogModel struct{}
