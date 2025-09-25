package model

type SaleResponse struct {
	Code         string                `json:"code,omitempty"`
	CustomerName string                `json:"customer_name,omitempty"`
	Status       string                `json:"status,omitempty"`
	CreatedAt    int64                 `json:"created_at,omitempty"`
	BranchName   string                `json:"branch_name,omitempty"`
	BranchID     uint                  `json:"branch_id,omitempty"`
	TotalQty     int                   `json:"total_qty,omitempty"`
	TotalPrice   float64               `json:"total_price,omitempty"`
	Items        []SaleItemResponse    `json:"items,omitempty"`
	Payments     []SalePaymentResponse `json:"payments,omitempty"`
}

type SaleItemResponse struct {
	Size        *SizeResponse    `json:"size"`
	Product     *ProductResponse `json:"product"`
	Qty         int              `json:"qty"`
	Price       float64          `json:"price"`
	IsCancelled int              `json:"is_cancelled"`
}

type SearchSaleRequest struct {
	BranchID *uint  `json:"branch_id"`
	Search   string `json:"search"`
	Status   string `json:"status"`
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}

type GetSaleRequest struct {
	Code string `json:"-" validate:"required"`
}

type CreateSaleRequest struct {
	CustomerName string                     `json:"customer_name" validate:"required"`
	BranchID     uint                       `json:"branch_id" validate:"required"`
	Details      []CreateSaleDetailRequest  `json:"details" validate:"required,dive"`
	Payments     []CreateSalePaymentRequest `json:"payments"`
	Debt         *CreateDebtRequest         `json:"debt"`
}

type CancelSaleRequest struct {
	Code string `json:"-" validate:"required"`
}
