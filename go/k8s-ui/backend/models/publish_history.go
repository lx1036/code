package models

import "time"

type ReleaseStatus int32

const (
	ReleaseFailure ReleaseStatus = iota
	ReleaseSuccess

	TableNamePublishHistory = "publish_history"
)

type PublishHistory struct {
	Id           int64         `gorm:"auto" json:"id,omitempty"`
	Type         PublishType   `gorm:"index;type(integer)" json:"type,omitempty"`
	ResourceId   int64         `gorm:"index" json:"resourceId,omitempty"`
	ResourceName string        `gorm:"size(128)" json:"resourceName,omitempty"`
	TemplateId   int64         `gorm:"index;column(template_id)" json:"templateId,omitempty"`
	Cluster      string        `gorm:"size(128)" json:"cluster,omitempty"`
	Status       ReleaseStatus `gorm:"type(integer)" json:"status,omitempty"`
	Message      string        `gorm:"type(text)" json:"message,omitempty"`
	User         string        `gorm:"size(128)" json:"user,omitempty"`
	CreateTime   *time.Time    `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
}

func (*PublishHistory) TableName() string {
	return TableNamePublishHistory
}

type publishHistoryModel struct{}
