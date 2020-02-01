package models

import "time"

const (
	TableNameIngressTemplate = "ingress_template"
)

type IngressTemplate struct {
	Id          int64    `gorm:"auto" json:"id,omitempty"`
	Name        string   `gorm:"size(128)" json:"name,omitempty"`
	Template    string   `gorm:"type(text)" json:"template,omitempty"`
	Ingress     *Ingress `gorm:"index;rel(fk);column(ingress_id)" json:"ingress,omitempty"`
	Description string   `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status    []*PublishStatus `gorm:"-" json:"status,omitempty"`
	IngressId int64            `gorm:"-" json:"ingressId,omitempty"`
}

func (*IngressTemplate) TableName() string {
	return TableNameIngressTemplate
}

type ingressTemplateModel struct{}
