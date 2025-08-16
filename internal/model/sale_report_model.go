package model

import "time"

type SearchSalesDailyReportRequest struct {
	BranchID *uint     `json:"-"`
	Search   string    `json:"-"`
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Page     int       `json:"page" validate:"min=1"`
	Size     int       `json:"size" validate:"min=1,max=100"`
}

type SalesDailyReportResponse struct {
	Date              string             `json:"date"`
	BranchID          uint               `json:"branch_id"`
	BranchName        string             `json:"branch_name"`
	TotalTransactions int                `json:"total_transactions"`
	TotalProductsSold int                `json:"total_products_sold"`
	TotalRevenue      float64            `json:"total_revenue"`
	PaymentMethods    map[string]float64 `json:"payment_methods"`
}
