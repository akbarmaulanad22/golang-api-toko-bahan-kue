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

// type Debt struct {
// 	ID            int       `json:"id"`
// 	ReferenceType string    `json:"reference_type"`
// 	ReferenceCode string    `json:"reference_code"`
// 	TotalAmount   float64   `json:"total_amount"`
// 	PaidAmount    float64   `json:"paid_amount"`
// 	DueDate       int64     `json:"due_date"` // Unix timestamp
// 	Status        string    `json:"status"`
// 	CreatedAt     time.Time `json:"created_at"`
// 	UpdatedAt     time.Time `json:"updated_at"`
// }

// type DebtPayment struct {
// 	ID          int       `json:"id"`
// 	DebtID      int       `json:"debt_id"`
// 	PaymentDate time.Time `json:"payment_date"`
// 	Amount      float64   `json:"amount"`
// 	Note        string    `json:"note,omitempty"`
// 	CreatedAt   time.Time `json:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at"`
// }

// type Sale struct {
// 	Code         string    `json:"code"`
// 	CustomerName string    `json:"customer_name"`
// 	Status       string    `json:"status"`
// 	BranchID     int       `json:"branch_id"`
// 	CreatedAt    time.Time `json:"created_at"`
// }

// type SaleDetail struct {
// 	SizeID      int     `json:"size_id"`
// 	ProductSKU  string  `json:"product_sku"`
// 	ProductName string  `json:"product_name"`
// 	Qty         int     `json:"qty"`
// 	SellPrice   float64 `json:"sell_price"`
// }

// func (d *CreateDebtRequest) UnmarshalJSON(data []byte) error {
// 	type Alias CreateDebtRequest
// 	aux := &struct {
// 		DueDate interface{} `json:"due_date"`
// 		*Alias
// 	}{
// 		Alias: (*Alias)(d),
// 	}

// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}

// 	switch v := aux.DueDate.(type) {
// 	case string:
// 		t, err := time.Parse("2006-01-02", v)
// 		if err != nil {
// 			return fmt.Errorf("invalid date format: %w", err)
// 		}
// 		d.DueDate = t.UnixMilli()
// 	case float64: // kalau user sudah kirim timestamp
// 		d.DueDate = int64(v)
// 	default:
// 		return fmt.Errorf("unsupported date type")
// 	}

// 	return nil
// }
