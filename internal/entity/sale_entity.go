package entity

import (
	"tokobahankue/internal/model"
)

type Sale struct {
	Code         string              `gorm:"column:code;primaryKey"`
	CustomerName string              `gorm:"column:customer_name"`
	Status       model.StatusPayment `gorm:"column:status"`
	CashValue    int                 `gorm:"column:cash_value"`
	DebitValue   int                 `gorm:"column:debit_value"`
	PaidAt       int64               `gorm:"column:paid_at"`
	BranchID     uint                `gorm:"column:branch_id"`
	CreatedAt    int64               `gorm:"column:created_at;autoCreateTime:milli"`
	CancelledAt  int64               `gorm:"column:cancelled_at"`

	Branch      Branch       `gorm:"foreignKey:BranchID;references:ID"`
	SaleDetails []SaleDetail `gorm:"foreignKey:SaleCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
