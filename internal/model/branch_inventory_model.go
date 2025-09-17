package model

type BranchInventorySizeResponse struct {
	SizeID    uint    `json:"size_id"`
	Size      string  `json:"size"`
	Stock     int     `json:"stock"`
	SellPrice float64 `json:"sell_price"`
}

type BranchInventoryProductResponse struct {
	BranchName  string                        `json:"branch_name,omitempty"`
	ProductSKU  string                        `json:"sku"`
	ProductName string                        `json:"name"`
	Sizes       []BranchInventorySizeResponse `json:"sizes"`
}

type SearchBranchInventoryRequest struct {
	BranchID *uint  `json:"-"`
	Search   string `json:"search" validate:"max=100"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}

// type BranchInventoryAdminRequest struct {
// 	BranchID uint `json:"-" validate:"required"`
// 	Page     int  `json:"page" validate:"min=1"`
// 	Size     int  `json:"size" validate:"min=1,max=	100"`
// }

// type CreateBranchInventoryRequest struct {
// 	Stock    int  `json:"stock" validate:"required"`
// 	BranchID uint `json:"-" validate:"required"`
// 	SizeID   uint `json:"size_id" validate:"required"`
// }

// type UpdateStockBranchInventoryRequest struct {
// 	ChangeQty         int  `json:"change_qty" validate:"required"`
// 	BranchInventoryID uint `json:"branch_inventory_id" validate:"required"`
// }

// type CreateBranchInventoryRequest struct {
// 	Name    string `json:"name" validate:"required,max=100"`
// 	Address string `json:"address" validate:"required,max=100"`
// }

// type SearchBranchInventoryRequest struct {
// 	Name    string `json:"name" validate:"max=100"`
// 	Address string `json:"address" validate:"max=100"`
// 	Page    int    `json:"page" validate:"min=1"`
// 	Size    int    `json:"size" validate:"min=1,max=100"`
// }

// type GetBranchInventoryRequest struct {
// 	ID uint `json:"id" validate:"required,max=100"`
// }

// type UpdateBranchRequest struct {
// 	ID      uint   `json:"-" validate:"required,max=100"`
// 	Name    string `json:"name,omitempty" validate:"max=100"`
// 	Address string `json:"address,omitempty" validate:"max=100"`
// }

// type DeleteBranchRequest struct {
// 	ID uint `json:"-" validate:"required,max=100"`
// }
