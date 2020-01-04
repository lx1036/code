package models

const (
	TableNameLinkType = "link_type"
)

type linkTypeModel struct{}

type LinkType struct {
	Id          int64  `orm:"auto" json:"id,omitempty"`
	TypeName    string `orm:"index;unique;size(255)" json:"typeName,omitempty"`
	Displayname string `orm:"size(255)" json:"displayname,omitempty"`
	DefaultUrl  string `orm:"size(255)" json:"defaultUrl,omitempty"`
	ParamList   string `orm:"size(255);null" json:"paramList,omitempty"`
	Deleted     bool   `orm:"default(false)" json:"deleted,omitempty"`
}

func (*LinkType) TableName() string {
	return TableNameLinkType
}
