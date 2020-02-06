package models

import "time"

const (
	TableNameCharge = "charges"
)

type ResourceName string

type Charge struct {
	ID           uint         `gorm:"column:id;primary_key;"`
	Namespace    string       `gorm:"column:namespace;size:1024;not null;default:'';"`
	App          string       `gorm:"column:app;size:128;not null;default:'';"`
	Name         string       `gorm:"column:name;size:1024;not null;default:'';"`
	Type         string       `gorm:"column:type;size:128;not null;default:'';"`
	UnitPrice    float64      `gorm:"column:unit_price;precision:4;not null;default:'0.0000'"`
	Quantity     int          `gorm:"column:quantity;size:11;not null;default:0;"`
	Amount       float64      `gorm:"column:amount;precision:4;not null;default:'0.0000'"`
	ResourceName ResourceName `gorm:"column:resource_name;size:1024;not null;default:'';"`
	CreatedAt    time.Time    `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt    time.Time    `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt    time.Time    `gorm:"column:deleted_at;default:null;"`
}

func (Charge) TableName() string {
	return TableNameCharge
}

type chargeModel struct{}
