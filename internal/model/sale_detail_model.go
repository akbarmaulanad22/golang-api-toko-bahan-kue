package model

type SaleDetailResponse struct {
	SizeID      uint          `json:"size_id,omitempty"`
	Qty         int           `json:"qty,omitempty"`
	SellPrice   float64       `json:"sell_price,omitempty"`
	IsCancelled bool          `json:"is_cancelled,omitempty"`
	Size        *SizeResponse `json:"size,omitempty"`
}

type CreateSaleDetailRequest struct {
	SizeID    uint    `json:"size_id" validate:"required"`
	Qty       int     `json:"qty" validate:"required,min=1"`
	SellPrice float64 `json:"-"`
}
