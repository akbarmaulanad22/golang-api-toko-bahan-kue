package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func ExpenseToResponse(e *entity.Expense) *model.ExpenseResponse {

	var branchName *string
	if e.Branch.Name != "" {
		branchName = &e.Branch.Name
	}

	return &model.ExpenseResponse{
		ID:          e.ID,
		Description: e.Description,
		Amount:      e.Amount,
		CreatedAt:   e.CreatedAt,
		BranchName:  branchName,
	}
}
