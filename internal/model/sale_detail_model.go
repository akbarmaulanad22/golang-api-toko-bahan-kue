package model

type SaleDetailResponse struct {
	SaleCode    string `json:"code,omitempty"`
	SizeID      uint   `json:"size_id,omitempty"`
	Qty         int    `json:"qty,omitempty"`
	IsCancelled bool   `json:"is_cancelled,omitempty"`

	Size SizeResponse `json:"size"`
}

type CreateSaleDetailRequest struct {
	SaleCode    string `json:"code" validate:"required,max=100"`
	SizeID      uint   `json:"size_id" validate:"required,max=100"`
	Qty         int    `json:"qty" validate:"required,max=100"`
	IsCancelled bool   `json:"-"`
}

type UpdateSaleDetailRequest struct {
	SaleCode    string `json:"-" validate:"required,max=100"`
	SizeID      uint   `json:"size_id" validate:"required,max=100"`
	IsCancelled bool   `json:"-"`
}
