package models

import "time"

const (
	TableNameSecretTemplate = "secret_template"
)

type secretTplModel struct{}

type SecretTemplate struct {
	Id       int64   `gorm:"auto" json:"id,omitempty"`
	Name     string  `gorm:"size(128)" json:"name,omitempty"`
	Template string  `gorm:"type(text)" json:"template,omitempty"`
	Secret   *Secret `gorm:"index;rel(fk);column(secret_map_id)" json:"secret,omitempty"`
	// 存储模版可上线机房
	// 例如{"clusters":["K8S"]}
	MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	Description string `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status   []*PublishStatus `gorm:"-" json:"status,omitempty"`
	SecretId int64            `gorm:"-" json:"secretId,omitempty"`
}

func (*SecretTemplate) TableName() string {
	return TableNameSecretTemplate
}
