package models

import "time"

type UserType int


type User struct {
	Id        int64      `orm:"pk;auto" json:"id,omitempty"`
	Name      string     `orm:"index;unique;size(200)" json:"name,omitempty"`
	Password  string     `orm:"size(255)" json:"-"`
	Salt      string     `orm:"size(32)" json:"-"`
	Email     string     `orm:"size(200)" json:"email,omitempty"`
	Display   string     `orm:"size(200)" json:"display,omitempty"`
	Comment   string     `orm:"type(text)" json:"comment,omitempty"`
	Type      UserType   `orm:"type(integer)" json:"type"`
	Admin     bool       `orm:"default(False)" json:"admin"`
	LastLogin *time.Time `orm:"auto_now_add;type(datetime)" json:"lastLogin,omitempty"`
	LastIp    string     `orm:"size(200)" json:"lastIp,omitempty"`

	Deleted    bool       `orm:"default(false)" json:"deleted,omitempty"`
	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`

	Namespaces []*Namespace `orm:"-" json:"namespaces,omitempty"`
}


