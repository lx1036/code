package models

import "time"

type ReleaseStatus int32

const (
	ReleaseFailure ReleaseStatus = iota
	ReleaseSuccess

	TableNamePublishHistory = "publish_history"
)

type PublishHistory struct {
	Id           int64         `orm:"auto" json:"id,omitempty"`
	Type         PublishType   `orm:"index;type(integer)" json:"type,omitempty"`
	ResourceId   int64         `orm:"index" json:"resourceId,omitempty"`
	ResourceName string        `orm:"size(128)" json:"resourceName,omitempty"`
	TemplateId   int64         `orm:"index;column(template_id)" json:"templateId,omitempty"`
	Cluster      string        `orm:"size(128)" json:"cluster,omitempty"`
	Status       ReleaseStatus `orm:"type(integer)" json:"status,omitempty"`
	Message      string        `orm:"type(text)" json:"message,omitempty"`
	User         string        `orm:"size(128)" json:"user,omitempty"`
	CreateTime   *time.Time    `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
}

func (*PublishHistory) TableName() string {
	return TableNamePublishHistory
}

type publishHistoryModel struct{}
