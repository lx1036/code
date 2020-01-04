package models

import "time"

const (
	TableNamePersistentVolumeClaimTemplate = "persistent_volume_claim_template"
)

type persistentVolumeClaimTplModel struct{}

type PersistentVolumeClaimTemplate struct {
	Id                    int64                  `orm:"auto" json:"id,omitempty"`
	Name                  string                 `orm:"size(128)" json:"name,omitempty"`
	Template              string                 `orm:"type(text)" json:"template,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaim `orm:"index;rel(fk)" json:"persistentVolumeClaim,omitempty"`
	// 存储模版可上线机房
	// 例如{"clusters":["K8S"]}
	MetaData    string `orm:"type(text)" json:"metaData,omitempty"`
	Description string `orm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `orm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `orm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `orm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `orm:"default(false)" json:"deleted,omitempty"`

	Status                  []*PublishStatus `orm:"-" json:"status,omitempty"`
	PersistentVolumeClaimId int64            `orm:"-" json:"persistentVolumeClaimId,omitempty"`
}

func (*PersistentVolumeClaimTemplate) TableName() string {
	return TableNamePersistentVolumeClaimTemplate
}
