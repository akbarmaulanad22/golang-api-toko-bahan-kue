package model

type SalePaymentResponse struct {
	PaymentMethod string  `json:"payment_method,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Note          string  `json:"note,omitempty"`
	CreatedAt     int64   `json:"created_at,omitempty"`
}

type CreateSalePaymentRequest struct {
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=CASH DEBIT TRANSFER QRIS EWALLET"`
	Amount        float64 `json:"amount" validate:"required"`
	Note          string  `json:"note" validate:"max=255"`
}
