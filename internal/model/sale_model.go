package model

type SaleResponse struct {
	Code         string                `json:"code,omitempty"`
	CustomerName string                `json:"customer_name,omitempty"`
	Status       string                `json:"status,omitempty"`
	CreatedAt    int64                 `json:"created_at,omitempty"`
	BranchID     uint                  `json:"branch_id,omitempty"`
	Details      []SaleDetailResponse  `json:"details,omitempty"`
	Payments     []SalePaymentResponse `json:"payments,omitempty"`
	Debt         *DebtResponse         `json:"debt,omitempty"`
}

type SearchSaleRequest struct {
	BranchID     uint   `json:"branch_id"`
	Code         string `json:"code" validate:"max=100"`
	CustomerName string `json:"customer_name" validate:"max=100"`
	Status       string `json:"status"`
	StartAt      int64  `json:"start_at"`
	EndAt        int64  `json:"end_at"`
	Page         int    `json:"page" validate:"min=1"`
	Size         int    `json:"size" validate:"min=1,max=100"`
}

type GetSaleRequest struct {
	Code string `json:"-" validate:"required,max=100"`
}

type CreateSaleRequest struct {
	CustomerName string                     `json:"customer_name" validate:"required,max=100"`
	BranchID     uint                       `json:"branch_id" validate:"required"`
	Details      []CreateSaleDetailRequest  `json:"details" validate:"required,dive"`
	Payments     []CreateSalePaymentRequest `json:"payments,omitempty"`
	Debt         *CreateDebtRequest         `json:"debt,omitempty"`
}

type CancelSaleRequest struct {
	Code string `json:"-" validate:"required,max=100"`
}

// type SaleReportResponse struct {
// 	CreatedAt    time.Time `json:"created_at"`
// 	BranchName   string    `json:"branch_name"`
// 	SaleCode     string    `json:"sale_code"`
// 	ProductName  string    `json:"product_name"`
// 	CustomerName string    `json:"customer_name"`
// 	Qty          int       `json:"qty"`
// 	SellPrice    float64   `json:"sell_price"`
// 	TotalPrice   float64   `json:"total_price"`
// }

// type SearchSaleReportRequest struct {
// 	BranchID uint   `json:"-"`
// 	Search   string `json:"-"`
// 	StartAt  string `json:"start_at"`
// 	EndAt    string `json:"end_at"`
// 	Page     int    `json:"page" validate:"min=1"`
// 	Size     int    `json:"size" validate:"min=1,max=100"`
// }

// type BranchSalesReportResponse struct {
// 	BranchName string  `json:"branch_name"`
// 	TotalSales float64 `json:"total_sales"`
// }

// type BestSellingProductResponse struct {
// 	ProductName string  `json:"product_name"`
// 	TotalQty    int64   `json:"total_qty"`
// 	TotalSales  float64 `json:"total_sales"`
// }

// type ListBestSellingProductRequest struct {
// 	BranchID uint `json:"-"`
// }
