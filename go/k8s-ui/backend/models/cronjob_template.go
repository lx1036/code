package models

import "time"

const (
	TableNameCronjobTemplate = "cronjob_template"
)

type CronjobTemplate struct {
	Id          int64    `gorm:"auto" json:"id,omitempty"`
	Name        string   `gorm:"size(128)" json:"name,omitempty"`
	Template    string   `gorm:"type(text)" json:"template,omitempty"`
	Cronjob     *Cronjob `gorm:"index;rel(fk);column(cronjob_id)" json:"cronjob,omitempty"`
	MetaData    string   `gorm:"type(text)" json:"metaData,omitempty"`
	Description string   `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status    []*PublishStatus `gorm:"-" json:"status,omitempty"`
	CronjobId int64            `gorm:"-" json:"cronjobId,omitempty"`
}

func (*CronjobTemplate) TableName() string {
	return TableNameCronjobTemplate
}

type cronjobTplModel struct{}
