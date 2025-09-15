package model

type CashBankTransactionResponse struct {
	ID              uint    `json:"id,omitempty"`
	TransactionDate int64   `json:"transaction_date,omitempty"`
	Type            string  `json:"type,omitempty"`
	Source          string  `json:"source,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
	Description     string  `json:"description,omitempty"`
	CreatedAt       int64   `json:"created_at,omitempty"`
	BranchName      string  `json:"branch_name,omitempty"`
}

type CreateCashBankTransactionRequest struct {
	TransactionDate int64   `json:"transaction_date" validate:"required"`
	Type            string  `json:"type" validate:"required"`
	Source          string  `json:"source" validate:"required"`
	Amount          float64 `json:"amount" validate:"required"`
	Description     string  `json:"description" validate:"required"`
	ReferenceKey    string  `json:"reference_key" validate:"required"`
	BranchID        uint    `json:"branch_id" validate:"required"`
}

type SearchCashBankTransactionRequest struct {
	StartAt int64   `json:"start_at"`
	EndAt   int64   `json:"end_at"`
	Amount  float64 `json:"amount"`
	// Type         string  `json:"type"`
	// Source       string  `json:"source"`
	// Description  string  `json:"description"`
	// ReferenceKey string  `json:"reference_key"`
	Page int `json:"page" validate:"min=1"`
	Size int `json:"size" validate:"min=1"`
}

type GetCashBankTransactionRequest struct {
	ID uint `json:"id" validate:"required"`
}

type UpdateCashBankTransactionRequest struct {
	ID              uint    `json:"-" validate:"required"`
	TransactionDate int64   `json:"transaction_date" validate:"required"`
	Type            string  `json:"type" validate:"required"`
	Source          string  `json:"source" validate:"required"`
	Amount          float64 `json:"amount" validate:"required"`
	Description     string  `json:"description" validate:"required"`
	ReferenceKey    string  `json:"reference_key" validate:"required"`
	// BranchID        uint    `json:"branch_id" validate:"required"`
}

type DeleteCashBankTransactionRequest struct {
	ID uint `json:"-" validate:"required"`
}
