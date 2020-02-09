package models

import "time"

const (
	TableNameLinkType = "link_types"
)

type linkTypeModel struct{}

type LinkType struct {
	ID          int64      `gorm:"column:id;primary_key;"`
	TypeName    string     `gorm:"column:type_name;size:255;not null;unique;default:'';"`
	DisplayName string     `gorm:"column:display_name;size:255;not null;default:'';"`
	DefaultUrl  string     `gorm:"column:default_url;size:255;not null;default:'';"`
	ParamList   string     `gorm:"column:param_list;size:255;default:null;"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;default:null;"`

	//Deleted     bool   `gorm:"default(false)" json:"deleted,omitempty"`
}

func (LinkType) TableName() string {
	return TableNameLinkType
}
