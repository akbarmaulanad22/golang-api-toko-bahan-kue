package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func CashBankTransactionToResponse(e *entity.CashBankTransaction) *model.CashBankTransactionResponse {
	return &model.CashBankTransactionResponse{
		ID:              e.ID,
		TransactionDate: e.TransactionDate,
		Type:            e.Type,
		Source:          e.Source,
		Amount:          e.Amount,
		Description:     e.Description,
		CreatedAt:       e.CreatedAt,
		BranchName:      e.Branch.Name,
	}
}
