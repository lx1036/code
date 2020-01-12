package models

import "time"

const (
	TableNameDaemonSetTemplate = "daemon_set_template"
)

type DaemonSetTemplate struct {
	Id          int64      `orm:"auto" json:"id,omitempty"`
	Name        string     `orm:"size(128)" json:"name,omitempty"`
	Template    string     `orm:"type(text)" json:"template,omitempty"`
	DaemonSet   *DaemonSet `orm:"index;rel(fk)" json:"daemonSet,omitempty"`
	Description string     `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	DaemonSetId int64            `orm:"-" json:"daemonSetId,omitempty"`
	Status      []*PublishStatus `orm:"-" json:"status,omitempty"`
}

func (*DaemonSetTemplate) TableName() string {
	return TableNameDaemonSetTemplate
}

type daemonSetTplModel struct{}
