package model

type PurchaseResponse struct {
	Code            string                    `json:"code,omitempty"`
	SalesName       string                    `json:"sales_name,omitempty"`
	Status          string                    `json:"status,omitempty"`
	CreatedAt       int64                     `json:"created_at,omitempty"`
	BranchName      string                    `json:"branch_name,omitempty"`
	DistributorName string                    `json:"distributor_name,omitempty"`
	TotalQty        int                       `json:"total_qty,omitempty"`
	TotalPrice      float64                   `json:"total_price,omitempty"`
	Items           []PurchaseItemResponse    `json:"items,omitempty"`
	Payments        []PurchasePaymentResponse `json:"payments,omitempty"`
}

type PurchaseItemResponse struct {
	Size        *SizeResponse    `json:"size"`
	Product     *ProductResponse `json:"product"`
	Qty         int              `json:"qty"`
	Price       float64          `json:"price"`
	IsCancelled int              `json:"is_cancelled"`
}

type SearchPurchaseRequest struct {
	BranchID *uint  `json:"branch_id"`
	Search   string `json:"search"`
	Status   string `json:"status"`
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}

type GetPurchaseRequest struct {
	Code string `json:"-" validate:"required"`
}

type CreatePurchaseRequest struct {
	SalesName     string                         `json:"sales_name" validate:"required"`
	BranchID      uint                           `json:"branch_id" validate:"required"`
	DistributorID uint                           `json:"distributor_id" validate:"required"`
	Details       []CreatePurchaseDetailRequest  `json:"details" validate:"required,dive"`
	Payments      []CreatePurchasePaymentRequest `json:"payments"`
	Debt          *CreateDebtRequest             `json:"debt"`
}

type CancelPurchaseRequest struct {
	Code string `json:"-" validate:"required"`
}
