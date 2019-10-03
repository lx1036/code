package models

import "time"

type Permission struct {
	Id      int64  `orm:"auto" json:"id,omitempty"`
	Name    string `orm:"index;size(200)" json:"name,omitempty"`
	Comment string `orm:"type(text)" json:"comment,omitempty"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Groups []*Group `orm:"reverse(many)" json:"groups,omitempty"`
}

