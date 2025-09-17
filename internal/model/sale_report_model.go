package model

type SearchSalesReportRequest struct {
	BranchID *uint `json:"-"`
	// Search   string    `json:"-"`
	StartAt int64 `json:"start_at"`
	EndAt   int64 `json:"end_at"`
	Page    int   `json:"page" validate:"min=1"`
	Size    int   `json:"size" validate:"min=1,max=100"`
}

// type SalesDailyReportResponse struct {
// 	Date              string             `json:"date"`
// 	BranchID          uint               `json:"branch_id"`
// 	BranchName        string             `json:"branch_name"`
// 	TotalTransactions int                `json:"total_transactions"`
// 	TotalProductsSold int                `json:"total_products_sold"`
// 	TotalRevenue      float64            `json:"total_revenue"`   // dari sale_details
// 	TotalPayment      float64            `json:"total_payment"`   // dari sale_payments
// 	TotalDebt         float64            `json:"total_debt"`      // dari debts
// 	Balance           float64            `json:"balance"`         // payment - revenue
// 	PaymentMethods    map[string]float64 `json:"payment_methods"` // breakdown sale_payments
// }

type SalesDailyReportResponse struct {
	Date              string  `json:"date"`
	BranchID          uint    `json:"branch_id"`
	BranchName        string  `json:"branch_name"`
	TotalTransactions int     `json:"total_transactions"`
	TotalProductsSold int     `json:"total_products_sold"`
	TotalRevenue      float64 `json:"total_revenue"`
}

type SalesTopSellerReportResponse struct {
	BranchName  string  `json:"branch_name"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	TotalQty    int64   `json:"total_qty"`
	TotalOmzet  float64 `json:"total_omzet"`
}

type SalesCategoryResponse struct {
	BranchName   string  `json:"branch_name"`
	CategoryName string  `json:"category_name"`
	TotalQty     int64   `json:"total_qty"`
	TotalOmzet   float64 `json:"total_omzet"`
}
