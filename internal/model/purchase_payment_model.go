package model

type PurchasePaymentResponse struct {
	PaymentMethod string  `json:"payment_method,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	Note          string  `json:"note,omitempty"`
	CreatedAt     int64   `json:"created_at,omitempty"`
}

type CreatePurchasePaymentRequest struct {
	PaymentMethod string  `json:"payment_method" validate:"required,oneof=CASH DEBIT TRANSFER QRIS EWALLET"`
	Amount        float64 `json:"amount" validate:"required"`
	Note          string  `json:"note,omitempty" validate:"max=255"`
}
