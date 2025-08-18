package model

type ExpenseResponse struct {
	ID          uint    `json:"id,omitempty"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	BranchID    uint    `json:"branch_id,omitempty"`
	CreatedAt   int64   `json:"created_at,omitempty"`
}

type CreateExpenseRequest struct {
	Description string  `json:"description" validate:"required,max=255"`
	Amount      float64 `json:"amount" validate:"required"`
	BranchID    uint    `json:"branch_id" validate:"required"`
}

type UpdateExpenseRequest struct {
	ID          uint    `json:"-" validate:"required"`
	Description string  `json:"description,omitempty" validate:"max=255"`
	Amount      float64 `json:"amount,omitempty"`
}

type SearchExpenseRequest struct {
	Description string `json:"description" validate:"max=100"`
	StartAt     int64  `json:"start_at"`
	EndAt       int64  `json:"end_at"`
	Page        int    `json:"page" validate:"min=1"`
	Size        int    `json:"size" validate:"min=1,max=100"`
}

type DeleteExpenseRequest struct {
	ID uint `json:"-" validate:"required"`
}
