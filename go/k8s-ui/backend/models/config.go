package models

const (
	TableNameConfig = "config"
)

type ConfigKey string

type Config struct {
	Id    int64     `gorm:"auto" json:"id,omitempty"`
	Name  ConfigKey `gorm:"size(256)" json:"name,omitempty"`
	Value string    `gorm:"type(text)" json:"value,omitempty"`
}

func (*Config) TableName() string {
	return TableNameConfig
}

type configModel struct{}
