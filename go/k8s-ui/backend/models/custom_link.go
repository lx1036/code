package models

import "time"

const (
	TableNameCustomLink = "custom_links"
)

type CustomLink struct {
	Id int64 `gorm:"column:id;primary_key;"`
	// namespace name
	Namespace string    `gorm:"column:namespace;size:255;not null;default:'';"`
	LinkType  string    `gorm:"column:link_type;size:255;not null;default:'';"`
	Url       string    `gorm:"column:url;size:255;not null;default:'';"`
	AddParam  bool      `gorm:"column:add_param;not null;default:0;"`
	Params    string    `gorm:"column:params;size:255;default:null;"`
	Status    bool      `gorm:"column:status;not null;default:1;"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:null;"`
}

func (CustomLink) TableName() string {
	return TableNameCustomLink
}

type customLinkModel struct{}
