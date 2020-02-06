package models

import "time"

const (
	TableNameConfigMapTemplate = "config_map_template"
)

type ConfigMapTemplate struct {
	Id        int64      `gorm:"auto" json:"id,omitempty"`
	Name      string     `gorm:"size(128)" json:"name,omitempty"`
	Template  string     `gorm:"type(text)" json:"template,omitempty"`
	ConfigMap *ConfigMap `gorm:"index;rel(fk);column(config_map_id)" json:"configMap,omitempty"`
	// 存储模版可上线机房
	// 例如{"clusters":["K8S"]}
	MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	Description string `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status      []*PublishStatus `gorm:"-" json:"status,omitempty"`
	ConfigMapId int64            `gorm:"-" json:"configMapId,omitempty"`
}

func (*ConfigMapTemplate) TableName() string {
	return TableNameConfigMapTemplate
}

type configMapTplModel struct{}
