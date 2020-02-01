package models

import "time"

const (
	TableNameDaemonSetTemplate = "daemon_set_template"
)

type DaemonSetTemplate struct {
	Id          int64      `gorm:"auto" json:"id,omitempty"`
	Name        string     `gorm:"size(128)" json:"name,omitempty"`
	Template    string     `gorm:"type(text)" json:"template,omitempty"`
	DaemonSet   *DaemonSet `gorm:"index;rel(fk)" json:"daemonSet,omitempty"`
	Description string     `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	DaemonSetId int64            `gorm:"-" json:"daemonSetId,omitempty"`
	Status      []*PublishStatus `gorm:"-" json:"status,omitempty"`
}

func (*DaemonSetTemplate) TableName() string {
	return TableNameDaemonSetTemplate
}

type daemonSetTplModel struct{}
