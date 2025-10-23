package model

type PurchaseDetailResponse struct {
	ID          uint             `json:"id"`
	Size        *SizeResponse    `json:"size"`
	Product     *ProductResponse `json:"product"`
	Qty         int              `json:"qty"`
	Price       float64          `json:"price"`
	IsCancelled int              `json:"is_cancelled"`
}

type CreatePurchaseDetailRequest struct {
	BranchInventoryID uint    `json:"branch_inventory_id" validate:"required"`
	Qty               int     `json:"qty" validate:"required,min=1"`
	BuyPrice          float64 `json:"buy_price" validate:"required"`
}

type CancelPurchaseDetailRequest struct {
	ID           uint   `json:"-" validate:"required"`
	PurchaseCode string `json:"-" validate:"required"`
}
