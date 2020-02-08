package models

import "time"

const (
	TableNameHPATemplate = "hpa_templates"
)

type HPATemplate struct {
	ID          uint       `gorm:"column:id;primary_key;"`
	Name        string     `gorm:"column:name;size:128;not null;default:'';"`
	Template    string     `gorm:"column:template;type:longtext;not null;"`
	HpaId       uint       `gorm:"column:hpa_id"`
	MetaData    string     `gorm:"column:meta_data;type:longtext;not null;"`
	Description string     `gorm:"column:description;size:512;not null;default:'';"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`
}

func (HPATemplate) TableName() string {
	return TableNameHPATemplate
}

type hpaTemplateModel struct{}
