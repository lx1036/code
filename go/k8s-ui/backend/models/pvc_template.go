package models

import "time"

const (
	TableNamePersistentVolumeClaimTemplate = "persistent_volume_claim_template"
)

type persistentVolumeClaimTplModel struct{}

type PersistentVolumeClaimTemplate struct {
	Id                    int64                  `gorm:"auto" json:"id,omitempty"`
	Name                  string                 `gorm:"size(128)" json:"name,omitempty"`
	Template              string                 `gorm:"type(text)" json:"template,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaim `gorm:"index;rel(fk)" json:"persistentVolumeClaim,omitempty"`
	// 存储模版可上线机房
	// 例如{"clusters":["K8S"]}
	MetaData    string `gorm:"type(text)" json:"metaData,omitempty"`
	Description string `gorm:"size(512)" json:"description,omitempty"`

	CreateTime time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
	UpdateTime time.Time `gorm:"auto_now;type(datetime)" json:"updateTime,omitempty"`
	User       string    `gorm:"size(128)" json:"user,omitempty"`
	Deleted    bool      `gorm:"default(false)" json:"deleted,omitempty"`

	Status                  []*PublishStatus `gorm:"-" json:"status,omitempty"`
	PersistentVolumeClaimId int64            `gorm:"-" json:"persistentVolumeClaimId,omitempty"`
}

func (*PersistentVolumeClaimTemplate) TableName() string {
	return TableNamePersistentVolumeClaimTemplate
}
