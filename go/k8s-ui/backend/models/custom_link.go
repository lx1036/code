package models

const (
	TableNameCustomLink = "custom_link"
)

type CustomLink struct {
	Id int64 `orm:"auto" json:"id,omitempty"`
	// namespace name
	Namespace string `orm:"index;namespace;" json:"namespace"`
	// LinkType typeName
	LinkType string `orm:"size(255)" json:"linkType,omitempty"`
	Url      string `orm:"size(255)" json:"url,omitempty"`
	AddParam bool   `orm:"default(false)" json:"addParam,omitempty"`
	Params   string `orm:"size(255);null" json:"params,omitempty"`
	Deleted  bool   `orm:"default(false)" json:"deleted,omitempty"`
	//链接状态，默认启用，false为禁用
	Status bool `orm:"default(true)" json:"status,omitempty"`

	Displayname string `orm:"-" json:"displayname,omitempty"`
}

func (*CustomLink) TableName() string {
	return TableNameCustomLink
}

type customLinkModel struct{}
