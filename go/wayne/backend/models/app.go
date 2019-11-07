package models

import "time"

type App struct {
	Id        int64      `orm:"auto" json:"id,omitempty"`
	Name      string     `orm:"index;size(128)" json:"name,omitempty"`
	Namespace *Namespace `orm:"index;column(namespace_id);rel(fk)" json:"namespace"`
	/*
		{
		    "mode": "beta",
		    "system.api-name-generate-rule":"none" // refers to models.Config ConfigKeyApiNameGenerateRule
		}
	*/
	MetaData    string `orm:"type(text)" json:"metaData,omitempty"`
	Description string `orm:"null;size(512)" json:"description,omitempty"`

	CreateTime *time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime *time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string     `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool       `orm:"default(false)" json:"deleted,omitempty"`

	// 用于权限的关联查询
	AppUsers []*AppUser `orm:"reverse(many)" json:"-"`

	// 关注的关联查询
	AppStars []*AppStarred `orm:"reverse(many)" json:"-"`
}
