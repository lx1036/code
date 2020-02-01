package models

import "time"

const (
	TableNameServiceTemplate = "service_template"
)

type ServiceTemplate struct {
	Id          int64    `gorm:"auto" json:"id,omitempty"`
	Name        string   `gorm:"size(128)" json:"name,omitempty"`
	Template    string   `gorm:"type(text)" json:"template,omitempty"`
	Service     *Service `gorm:"index;rel(fk);column(service_id)" json:"service,omitempty"`
	Description string   `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status    []*PublishStatus `gorm:"-" json:"status,omitempty"`
	ServiceId int64            `gorm:"-" json:"serviceId,omitempty"`
}

func (*ServiceTemplate) TableName() string {
	return TableNameServiceTemplate
}
