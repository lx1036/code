package models

import "time"

const (
	TableNameInvoice = "invoices"
)

type Invoice struct {
	ID        uint      `gorm:"column:id;primary_key;"`
	Namespace string    `gorm:"column:namespace;size:1024;not null;default:'';"`
	App       string    `gorm:"column:app;size:128;not null;default:'';"`
	Amount    float64   `gorm:"column:amount;precision:4;not null;default:'0.0000';"`
	EndedAt   time.Time `gorm:"column:ended_at;not null;default:current_timestamp;"`
	BilledAt  time.Time `gorm:"column:billed_at;not null;default:current_timestamp;"`
	StartedAt time.Time `gorm:"column:started_at;not null;default:current_timestamp;"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:current_timestamp;"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;default:current_timestamp on update current_timestamp;"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:null;"`
	//StartDate *time.Time `gorm:"type(datetime)" json:"startDate,omitempty"`
	//EndDate   *time.Time `gorm:"type(datetime)" json:"endDate,omitempty"`
	//BillDate  *time.Time `gorm:"type(datetime)" json:"billDate,omitempty"`
	//
	//CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
}

func (Invoice) TableName() string {
	return TableNameInvoice
}

type invoiceModel struct{}
