package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func DebtToResponse(debt *entity.Debt) *model.DebtResponse {
	payments := make([]model.DebtPaymentResponse, len(debt.Payments))
	for i, payment := range debt.Payments {
		payments[i] = *DebtPaymentToResponse(&payment)
	}

	return &model.DebtResponse{
		ID:            debt.ID,
		ReferenceType: debt.ReferenceType,
		ReferenceCode: debt.ReferenceCode,
		TotalAmount:   debt.TotalAmount,
		PaidAmount:    debt.PaidAmount,
		DueDate:       debt.DueDate,
		Status:        debt.Status,
		Payments:      payments,
	}
}
