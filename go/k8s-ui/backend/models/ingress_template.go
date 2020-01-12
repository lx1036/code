package models

import "time"

const (
	TableNameIngressTemplate = "ingress_template"
)

type IngressTemplate struct {
	Id          int64    `orm:"auto" json:"id,omitempty"`
	Name        string   `orm:"size(128)" json:"name,omitempty"`
	Template    string   `orm:"type(text)" json:"template,omitempty"`
	Ingress     *Ingress `orm:"index;rel(fk);column(ingress_id)" json:"ingress,omitempty"`
	Description string   `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	Status    []*PublishStatus `orm:"-" json:"status,omitempty"`
	IngressId int64            `orm:"-" json:"ingressId,omitempty"`
}

func (*IngressTemplate) TableName() string {
	return TableNameIngressTemplate
}

type ingressTemplateModel struct{}
