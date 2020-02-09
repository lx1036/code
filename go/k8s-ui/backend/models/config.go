package models

const (
	TableNameConfig = "configs"
)

type ConfigKey string

type Config struct {
	ID    uint      `gorm:"column:id;primary_key;"`
	Name  ConfigKey `gorm:"column:name;size:256;not null;default:'';"`
	Value string    `gorm:"column:value;type:longtext;not null;"`
}

func (Config) TableName() string {
	return TableNameConfig
}

type configModel struct{}
