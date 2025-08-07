package model

type BranchSalesSummaryResponse struct {
	BranchID         uint   `json:"branch_id"`
	BranchName       string `json:"branch_name"`
	TotalSales       int    `json:"total_sales"`
	TotalProductSold int    `json:"total_product_sold"`
}

type DailySalesSummaryResponse struct {
	Date             string `json:"date"`
	TotalSales       int    `json:"total_sales"`
	TotalProductSold int    `json:"total_product_sold"`
}

type ListDailySalesSummaryRequest struct {
	BranchID uint   `json:"branch_id" validate:"required"`
	StartAt  string `json:"start_at"`
	EndAt    string `json:"end_at"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}
