package model

import "database/sql/driver"

type StatusPayment string

const (
	PENDING   StatusPayment = "PENDING"
	COMPLETED StatusPayment = "COMPLETED"
	CANCELLED StatusPayment = "CANCELLED"
)

func (ct *StatusPayment) Scan(value interface{}) error {
	*ct = StatusPayment(value.([]byte))
	return nil
}

func (ct StatusPayment) Value() (driver.Value, error) {
	return string(ct), nil
}

type SaleResponse struct {
	Code         string        `json:"code,omitempty"`
	CustomerName string        `json:"customer_name,omitempty"`
	Status       StatusPayment `json:"status,omitempty"`
	CashValue    int           `json:"cash_value,omitempty"`
	DebitValue   int           `json:"debit_value,omitempty"`
	PaidAt       int64         `json:"paid_at,omitempty"`
	CreatedAt    int64         `json:"created_at,omitempty"`
	CancelledAt  int64         `json:"cancelled_at,omitempty"`

	Branch      BranchResponse       `json:"branch"`
	SaleDetails []SaleDetailResponse `json:"sale_details"`
}

type CreateSaleRequest struct {
	CustomerName string                    `json:"customer_name" validate:"required,max=100"`
	Status       StatusPayment             `json:"status" validate:"required,oneof=PENDING COMPLETED CANCELLED"`
	CashValue    int                       `json:"cash_value" validate:"required"`
	DebitValue   int                       `json:"debit_value" validate:"required"`
	PaidAt       int64                     `json:"-"`
	BranchID     uint                      `json:"branch_id" validate:"required,max=100"`
	SaleDetails  []CreateSaleDetailRequest `json:"sale_details" validate:"required"`
}

type SearchSaleRequest struct {
	Code         string        `json:"code" validate:"max=100"`
	CustomerName string        `json:"customer_name" validate:"max=100"`
	Status       StatusPayment `json:"status" validate:"oneof=PENDING COMPLETED CANCELLED"`
	StartAt      int64         `json:"start_at"`
	EndAt        int64         `json:"end_at"`
	Page         int           `json:"page" validate:"min=1"`
	Size         int           `json:"size" validate:"min=1,max=100"`
}

type GetSaleRequest struct {
	Code string `json:"-" validate:"required,max=100"`
}

type UpdateSaleRequest struct {
	Code   string        `json:"-" validate:"required,max=100"`
	Status StatusPayment `json:"status" validate:"required,oneof=PENDING COMPLETED CANCELLED"`
}
