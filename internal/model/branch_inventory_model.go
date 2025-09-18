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
