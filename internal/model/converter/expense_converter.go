package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func ExpenseToResponse(e *entity.Expense) *model.ExpenseResponse {
	return &model.ExpenseResponse{
		ID:          e.ID,
		Description: e.Description,
		Amount:      e.Amount,
		BranchID:    e.BranchID,
		CreatedAt:   e.CreatedAt,
	}
}
