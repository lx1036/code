package models

import "time"

const (
	TableNameNotification = "notifications"
)

type NotificationType string

type NotificationLevel int

const (
	LowNotification NotificationLevel = iota
	MiddleNotification
	HighNotification
)

type Notification struct {
	ID      uint             `gorm:"column:id;primary_key;" json:"id"`
	Type    NotificationType `gorm:"column:type;size:128;not null;default:'';" json:"type"`
	Title   string           `gorm:"column:title;size:2000;not null;default:'';" json:"title"`
	Message string           `gorm:"column:message;type:longtext;not null;" json:"message"`
	FromUserId  uint              `gorm:"column:from_user_id" json:"from_user_id"`
	Level       NotificationLevel `gorm:"column:level;size:11;not null;default:'0';" json:"level"`
	IsPublished bool              `gorm:"column:is_published;not null;default:'0';" json:"is_published"`
	CreatedAt   time.Time         `gorm:"column:created_at;not null;default:current_timestamp;" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;" json:"updated_at"`
	DeletedAt   *time.Time        `gorm:"column:deleted_at;default:null;" json:"deleted_at"`

	NotificationLogs []NotificationLog `gorm:"foreignkey:NotificationID;association_foreignkey:ID;" json:"notification_logs,omitempty"`
	User             User              `gorm:"column:user;foreignkey:FromUserId;" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return TableNameNotification
}

type notificationModel struct{}
