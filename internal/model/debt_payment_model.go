package model

type DebtPaymentResponse struct {
	Amount      float64 `json:"amount,omitempty"`
	PaymentDate int64   `json:"payment_date,omitempty"`
	Note        string  `json:"note,omitempty"`
}

type CreateDebtPaymentRequest struct {
	DebtID      int     `json:"-" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	PaymentDate int64   `json:"-" validate:"required"`
	Note        string  `json:"note,omitempty" validate:"max=255"`
}
