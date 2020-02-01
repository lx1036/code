package models

import "time"

const (
	TableNameAuditLog = "audit_log"
)

type auditLogModel struct{}

type AuditLogLevel string

type AuditLogType string

type AuditLog struct {
	Id         int64         `gorm:"auto" json:"id,omitempty"`
	SubjectId  int64         `gorm:"type(bigint)" json:"subjectId,omitempty"`
	LogType    AuditLogType  `gorm:"index;size(128)" json:"logType,omitempty"`
	LogLevel   AuditLogLevel `gorm:"index;size(128)" json:"logLevel,omitempty"`
	Action     string        `gorm:"index;size(255)" json:"action,omitempty"`
	Message    string        `gorm:"type(text);null" json:"message,omitempty"`
	UserIp     string        `gorm:"size(200)" json:"userIp,omitempty"`
	User       string        `gorm:"index;size(128)" json:"user,omitempty"`
	CreateTime *time.Time    `gorm:"auto_now_add;type(datetime);null" json:"createTime,omitempty"`
}

func (*AuditLog) TableName() string {
	return TableNameAuditLog
}
