package model

type PurchaseDetailResponse struct {
	SizeID      uint          `json:"size_id,omitempty"`
	Qty         int           `json:"qty,omitempty"`
	BuyPrice    float64       `json:"buy_price,omitempty"`
	IsCancelled bool          `json:"is_cancelled,omitempty"`
	Size        *SizeResponse `json:"size,omitempty"`
}

type CreatePurchaseDetailRequest struct {
	SizeID   uint    `json:"size_id" validate:"required"`
	Qty      int     `json:"qty" validate:"required,min=1"`
	BuyPrice float64 `json:"buy_price" validate:"required"`
}

type CancelPurchaseDetailRequest struct {
	SizeID       uint   `json:"-" validate:"required"`
	PurchaseCode string `json:"-" validate:"required"`
}
