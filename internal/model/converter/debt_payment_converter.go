package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func DebtPaymentToResponse(debtPayment *entity.DebtPayment) *model.DebtPaymentResponse {
	return &model.DebtPaymentResponse{
		Amount:      debtPayment.Amount,
		PaymentDate: debtPayment.PaymentDate,
		Note:        debtPayment.Note,
	}
}
