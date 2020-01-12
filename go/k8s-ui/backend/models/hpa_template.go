package models

import "time"

const (
	TableNameHPATemplate = "hpa_template"
)

type HPATemplate struct {
	Id          int64            `orm:"auto" json:"id,omitempty"`
	Name        string           `orm:"size(128)" json:"name,omitempty"`
	Template    string           `orm:"type(text)" json:"template,omitempty"`
	HPA         *HPA             `orm:"index;rel(fk);column(hpa_id)" json:"hpa,omitempty"`
	Description string           `orm:"size(512)" json:"description,omitempty"`
	CreateTime  time.Time        `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  time.Time        `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User        string           `orm:"size(128)" json:"user,omitempty"`
	Deleted     bool             `orm:"default(false)" json:"deleted,omitempty"`
	Status      []*PublishStatus `orm:"-" json:"status,omitempty"`
	HPAId       int64            `orm:"-" json:"hpaId,omitempty"`
}

func (*HPATemplate) TableName() string {
	return TableNameHPATemplate
}

type hpaTemplateModel struct{}
