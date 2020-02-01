package models

import "time"

const (
	TableNameHPATemplate = "hpa_template"
)

type HPATemplate struct {
	Id          int64            `gorm:"auto" json:"id,omitempty"`
	Name        string           `gorm:"size(128)" json:"name,omitempty"`
	Template    string           `gorm:"type(text)" json:"template,omitempty"`
	HPA         *HPA             `gorm:"index;rel(fk);column(hpa_id)" json:"hpa,omitempty"`
	Description string           `gorm:"size(512)" json:"description,omitempty"`
	CreateTime  time.Time        `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime  time.Time        `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User        string           `gorm:"size(128)" json:"user,omitempty"`
	Deleted     bool             `gorm:"default(false)" json:"deleted,omitempty"`
	Status      []*PublishStatus `gorm:"-" json:"status,omitempty"`
	HPAId       int64            `gorm:"-" json:"hpaId,omitempty"`
}

func (*HPATemplate) TableName() string {
	return TableNameHPATemplate
}

type hpaTemplateModel struct{}
