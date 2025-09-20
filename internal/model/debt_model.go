package model

type DebtResponse struct {
	ID            uint     `json:"id,omitempty"`
	ReferenceType string   `json:"reference_type,omitempty"`
	ReferenceCode string   `json:"reference_code,omitempty"`
	TotalAmount   float64  `json:"total_amount,omitempty"`
	PaidAmount    float64  `json:"paid_amount,omitempty"`
	DueDate       UnixDate `json:"due_date,omitempty"`
	Status        string   `json:"status,omitempty"`
	Related       string   `json:"related,omitempty"`
	BranchName    string   `json:"branch_name,omitempty"`
	CreatedAt     int64    `json:"created_at,omitempty"`
	UpdatedAt     int64    `json:"updated_at,omitempty"`
}

type CreateDebtRequest struct {
	// baru
	// ReferenceType string `json:"reference_type" validate:"oneof=SALE PURCHASE"`
	// ReferenceCode string `json:"reference_code"`
	// Status        string  `json:"-" validate:"oneof=PENDING PAID VOID"`
	// TotalAmount   float64 `json:"-"`
	// lama
	DueDate      UnixDate                   `json:"due_date" validate:"required"`
	DebtPayments []CreateDebtPaymentRequest `json:"debt_payments,omitempty"`
}

type SearchDebtRequest struct {
	BranchID      *uint    `json:"branch_id"`
	ReferenceType string   `json:"reference_type"`
	Status        string   `json:"status"`
	Search        string   `json:"search"` // ref code, total amount, paid amount
	StartAt       UnixDate `json:"start_at"`
	EndAt         UnixDate `json:"end_at"`
	Page          int      `json:"page" validate:"min=1"`
	Size          int      `json:"size" validate:"min=1,max=100"`
}

type GetDebtRequest struct {
	ID uint `json:"-" validate:"required"`
}

type DebtDetailResponse struct {
	ID            uint                  `json:"id"`
	ReferenceType string                `json:"reference_type"`
	ReferenceCode string                `json:"reference_code"`
	TotalAmount   float64               `json:"total_amount"`
	PaidAmount    float64               `json:"paid_amount"`
	DueDate       int64                 `json:"due_date"`
	Status        string                `json:"status"`
	CreatedAt     int64                 `json:"created_at"`
	Payments      []DebtPaymentResponse `json:"payments" gorm:"-"`
	Items         []DebtItemResponse    `json:"items" gorm:"-"`
}

type DebtItemResponse struct {
	ProductName string  `json:"product_name"`
	SizeName    string  `json:"size_name"`
	Qty         int     `json:"qty"`
	SellPrice   float64 `json:"sell_price"`
}
