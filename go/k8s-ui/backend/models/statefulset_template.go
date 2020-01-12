package models

import "time"

const (
	TableNameStatefulsetTemplate = "statefulset_template"
)

type statefulsetTplModel struct{}

type StatefulsetTemplate struct {
	Id          int64        `orm:"auto" json:"id,omitempty"`
	Name        string       `orm:"size(128)" json:"name,omitempty"`
	Template    string       `orm:"type(text)" json:"template,omitempty"`
	Statefulset *Statefulset `orm:"index;rel(fk);column(statefulset_id)" json:"statefulset,omitempty"`
	Description string       `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	StatefulsetId int64            `orm:"-" json:"statefulsetId,omitempty"`
	Status        []*PublishStatus `orm:"-" json:"status,omitempty"`
}

func (*StatefulsetTemplate) TableName() string {
	return TableNameStatefulsetTemplate
}
