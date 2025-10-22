package model

type SaleDetailResponse struct {
	ID uint `json:"id"`
	// BranchInventoryID uint             `json:"branch_inventory_id"`
	Size        *SizeResponse    `json:"size"`
	Product     *ProductResponse `json:"product"`
	Qty         int              `json:"qty"`
	Price       float64          `json:"price"`
	IsCancelled int              `json:"is_cancelled"`
}

type CreateSaleDetailRequest struct {
	BranchInventoryID uint    `json:"branch_inventory_id" validate:"required"`
	Qty               int     `json:"qty" validate:"required,min=1"`
	SellPrice         float64 `json:"-"`
}

type CancelSaleDetailRequest struct {
	ID       uint   `json:"-" validate:"required"`
	SaleCode string `json:"-" validate:"required"`
}
