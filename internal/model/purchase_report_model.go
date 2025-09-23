package model

type SearchPurchasesReportRequest struct {
	BranchID *uint `json:"-"`
	StartAt  int64 `json:"start_at"`
	EndAt    int64 `json:"end_at"`
	Page     int   `json:"page" validate:"min=1"`
	Size     int   `json:"size" validate:"min=1,max=100"`
}

type PurchasesDailyReportResponse struct {
	Date              string  `json:"date"`
	BranchID          uint    `json:"branch_id"`
	BranchName        string  `json:"branch_name"`
	TotalTransactions int     `json:"total_transactions"`
	TotalProductsBuy  int     `json:"total_products_buy"`
	TotalPurchases    float64 `json:"total_purchases"`
}
