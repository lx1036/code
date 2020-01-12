package models

import "time"

const (
	TableNameConfigMapTemplate = "config_map_template"
)

type ConfigMapTemplate struct {
	Id        int64      `orm:"auto" json:"id,omitempty"`
	Name      string     `orm:"size(128)" json:"name,omitempty"`
	Template  string     `orm:"type(text)" json:"template,omitempty"`
	ConfigMap *ConfigMap `orm:"index;rel(fk);column(config_map_id)" json:"configMap,omitempty"`
	// 存储模版可上线机房
	// 例如{"clusters":["K8S"]}
	MetaData    string `orm:"type(text)" json:"metaData,omitempty"`
	Description string `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	Status      []*PublishStatus `orm:"-" json:"status,omitempty"`
	ConfigMapId int64            `orm:"-" json:"configMapId,omitempty"`
}

func (*ConfigMapTemplate) TableName() string {
	return TableNameConfigMapTemplate
}

type configMapTplModel struct{}
