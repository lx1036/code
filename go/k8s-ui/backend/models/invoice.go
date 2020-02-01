package models

import "time"

const (
	TableNameInvoice = "invoice"
)

type Invoice struct {
	Id        int64  `gorm:"auto" json:"id,omitempty"`
	Namespace string `gorm:"size(1024)" json:"namespace,omitempty"`
	App       string `gorm:"index;size(128)" json:"app,omitempty"`

	Amount float64 `gorm:"digits(12);decimals(4)" json:"amount,omitempty"`

	StartDate *time.Time `gorm:"type(datetime)" json:"startDate,omitempty"`
	EndDate   *time.Time `gorm:"type(datetime)" json:"endDate,omitempty"`
	BillDate  *time.Time `gorm:"type(datetime)" json:"billDate,omitempty"`

	CreateTime *time.Time `gorm:"auto_now_add;type(datetime)" json:"createTime,omitempty"`
}

func (*Invoice) TableName() string {
	return TableNameInvoice
}

type invoiceModel struct{}
