package models

import "time"

const (
	TableNameCharge = "charge"
)

type ResourceName string

type Charge struct {
	Id        int64  `gorm:"auto" json:"id,omitempty"`
	Namespace string `gorm:"size(1024)" json:"namespace,omitempty"`
	App       string `gorm:"index;size(128)" json:"app,omitempty"`
	Name      string `gorm:"size(1024)" json:"name,omitempty"`
	Type      string `gorm:"index;size(128)" json:"type,omitempty"`

	UnitPrice float64 `gorm:"digits(12);decimals(4)" json:"unitPrice,omitempty"`
	Quantity  int     `gorm:"int(11)" json:"quantity,omitempty"`
	Amount    float64 `gorm:"digits(12);decimals(4)" json:"amount,omitempty"`

	ResourceName ResourceName `gorm:"size(1024)" json:"resourceName,omitempty"`
	StartTime    *time.Time   `gorm:"type(datetime)" json:"startTime,omitempty"`
	CreateTime   *time.Time   `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
}

func (*Charge) TableName() string {
	return TableNameCharge
}

type chargeModel struct{}
