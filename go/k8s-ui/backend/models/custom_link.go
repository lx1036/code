package models

const (
	TableNameCustomLink = "custom_link"
)

type CustomLink struct {
	Id int64 `gorm:"auto" json:"id,omitempty"`
	// namespace name
	Namespace string `gorm:"index;namespace;" json:"namespace"`
	// LinkType typeName
	LinkType string `gorm:"size(255)" json:"linkType,omitempty"`
	Url      string `gorm:"size(255)" json:"url,omitempty"`
	AddParam bool   `gorm:"default(false)" json:"addParam,omitempty"`
	Params   string `gorm:"size(255);null" json:"params,omitempty"`
	Deleted  bool   `gorm:"default(false)" json:"deleted,omitempty"`
	//链接状态，默认启用，false为禁用
	Status bool `gorm:"default(true)" json:"status,omitempty"`

	Displayname string `gorm:"-" json:"displayname,omitempty"`
}

func (*CustomLink) TableName() string {
	return TableNameCustomLink
}

type customLinkModel struct{}
