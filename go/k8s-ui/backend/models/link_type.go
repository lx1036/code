package models

const (
	TableNameLinkType = "link_type"
)

type linkTypeModel struct{}

type LinkType struct {
	Id          int64  `gorm:"auto" json:"id,omitempty"`
	TypeName    string `gorm:"index;unique;size(255)" json:"typeName,omitempty"`
	Displayname string `gorm:"size(255)" json:"displayname,omitempty"`
	DefaultUrl  string `gorm:"size(255)" json:"defaultUrl,omitempty"`
	ParamList   string `gorm:"size(255);null" json:"paramList,omitempty"`
	Deleted     bool   `gorm:"default(false)" json:"deleted,omitempty"`
}

func (*LinkType) TableName() string {
	return TableNameLinkType
}
