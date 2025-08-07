package entity

import (
	"tokobahankue/internal/model"
)

type Purchase struct {
	Code          string              `gorm:"column:code;primaryKey"`
	SalesName     string              `gorm:"column:sales_name"`
	Status        model.StatusPayment `gorm:"column:status"`
	CashValue     int                 `gorm:"column:cash_value"`
	DebitValue    int                 `gorm:"column:debit_value"`
	PaidAt        int64               `gorm:"column:paid_at"`
	BranchID      uint                `gorm:"column:branch_id"`
	DistributorID uint                `gorm:"column:distributor_id"`
	CreatedAt     int64               `gorm:"column:created_at;autoCreateTime:milli"`
	CancelledAt   int64               `gorm:"column:cancelled_at"`

	Branch          Branch           `gorm:"foreignKey:BranchID;references:ID"`
	Distributor     Distributor      `gorm:"foreignKey:DistributorID;references:ID"`
	PurchaseDetails []PurchaseDetail `gorm:"foreignKey:PurchaseCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
