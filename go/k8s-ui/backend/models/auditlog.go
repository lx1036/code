package models

import "time"

const (
	TableNameAuditLog = "audit_logs"
)

type auditLogModel struct{}

type AuditLogLevel string

type AuditLogType string

type AuditLog struct {
	ID        int64         `gorm:"column:id;primary_key;"`
	SubjectId int64         `gorm:"column:subject_id;size:20;not null;default:0;"`
	LogType   AuditLogType  `gorm:"column:log_type;size:128;not null;default:'';"`
	LogLevel  AuditLogLevel `gorm:"column:log_level;size:128;not null;default:'';"`
	Action    string        `gorm:"column:action;size:255;not null;default:'';"`
	Message   string        `gorm:"column:message;type:longtext;default:null;"`
	UserIp    string        `gorm:"column:user_ip;size:200;not null;default:'';"`
	User      string        `gorm:"column:user;size:128;not null;default:'';"`
	CreatedAt time.Time     `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt time.Time     `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt *time.Time    `gorm:"column:deleted_at;default:null;"`
}

func (AuditLog) TableName() string {
	return TableNameAuditLog
}
