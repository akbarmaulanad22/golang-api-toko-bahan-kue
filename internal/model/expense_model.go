package model

type ExpenseResponse struct {
	ID          int     `json:"id,omitempty"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	BranchID    int     `json:"branch_id,omitempty"`
	CreatedAt   int64   `json:"created_at,omitempty"`
}

type CreateExpenseRequest struct {
	Description string  `json:"description" validate:"required,max=255"`
	Amount      float64 `json:"amount" validate:"required"`
	BranchID    int     `json:"branch_id" validate:"required"`
}

type UpdateExpenseRequest struct {
	ID          int     `json:"-" validate:"required"`
	Description string  `json:"description,omitempty" validate:"max=255"`
	Amount      float64 `json:"amount,omitempty"`
}
