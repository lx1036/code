package models

const (
	TableNameConfig = "config"
)

type ConfigKey string

type Config struct {
	Id    int64     `orm:"auto" json:"id,omitempty"`
	Name  ConfigKey `orm:"size(256)" json:"name,omitempty"`
	Value string    `orm:"type(text)" json:"value,omitempty"`
}

func (*Config) TableName() string {
	return TableNameConfig
}

type configModel struct{}
