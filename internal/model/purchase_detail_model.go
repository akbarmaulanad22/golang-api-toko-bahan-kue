package model

type PurchaseDetailResponse struct {
	PurchaseCode string `json:"code,omitempty"`
	SizeID       uint   `json:"size_id,omitempty"`
	Qty          int    `json:"qty,omitempty"`
	IsCancelled  bool   `json:"is_cancelled,omitempty"`

	Size SizeResponse `json:"size"`
}

type CreatePurchaseDetailRequest struct {
	PurchaseCode string `json:"code" validate:"required,max=100"`
	SizeID       uint   `json:"size_id" validate:"required,max=100"`
	Qty          int    `json:"qty" validate:"required,max=100"`
	IsCancelled  bool   `json:"-"`
}

type UpdatePurchaseDetailRequest struct {
	PurchaseCode string `json:"-" validate:"required,max=100"`
	SizeID       uint   `json:"size_id" validate:"required,max=100"`
	IsCancelled  bool   `json:"-"`
}
