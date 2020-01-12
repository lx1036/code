package models

import "time"

const (
	TableNameCronjobTemplate = "cronjob_template"
)

type CronjobTemplate struct {
	Id          int64    `orm:"auto" json:"id,omitempty"`
	Name        string   `orm:"size(128)" json:"name,omitempty"`
	Template    string   `orm:"type(text)" json:"template,omitempty"`
	Cronjob     *Cronjob `orm:"index;rel(fk);column(cronjob_id)" json:"cronjob,omitempty"`
	MetaData    string   `orm:"type(text)" json:"metaData,omitempty"`
	Description string   `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	Status    []*PublishStatus `orm:"-" json:"status,omitempty"`
	CronjobId int64            `orm:"-" json:"cronjobId,omitempty"`
}

func (*CronjobTemplate) TableName() string {
	return TableNameCronjobTemplate
}

type cronjobTplModel struct{}
