package models

import "time"

const (
	TableNameStatefulsetTemplate = "statefulset_template"
)

type statefulsetTplModel struct{}

type StatefulsetTemplate struct {
	Id          int64        `gorm:"auto" json:"id,omitempty"`
	Name        string       `gorm:"size(128)" json:"name,omitempty"`
	Template    string       `gorm:"type(text)" json:"template,omitempty"`
	Statefulset *Statefulset `gorm:"index;rel(fk);column(statefulset_id)" json:"statefulset,omitempty"`
	Description string       `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	StatefulsetId int64            `gorm:"-" json:"statefulsetId,omitempty"`
	Status        []*PublishStatus `gorm:"-" json:"status,omitempty"`
}

func (*StatefulsetTemplate) TableName() string {
	return TableNameStatefulsetTemplate
}
