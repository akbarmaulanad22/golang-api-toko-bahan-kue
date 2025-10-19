package model

type StockOpnameDetailResponse struct {
	ID                uint   `json:"id"`
	StockOpnameID     uint   `json:"stock_opname_id"`
	BranchInventoryID uint   `json:"branch_inventory_id"`
	SystemQty         int64  `json:"system_qty"`
	PhysicalQty       int64  `json:"physical_qty"`
	Difference        int64  `json:"difference"`
	Notes             string `json:"notes"`
}

type CreateStockOpnameDetailInput struct {
	SizeID      uint   `json:"size_id" validate:"required"`
	PhysicalQty int64  `json:"physical_qty" validate:"min=0"`
	Notes       string `json:"notes" validate:"max=255"`
}

type UpdateStockOpnameDetailRequest struct {
	ID          uint   `json:"id" validate:"required"`
	PhysicalQty int64  `json:"physical_qty" validate:"min=0"`
	Notes       string `json:"notes" validate:"max=255"`
}
