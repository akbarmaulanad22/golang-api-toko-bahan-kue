package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SalePaymentToResponse(sale *entity.SalePayment) *model.SalePaymentResponse {
	return &model.SalePaymentResponse{
		PaymentMethod: sale.PaymentMethod,
		Amount:        sale.Amount,
		Note:          sale.Note,
		CreatedAt:     sale.CreatedAt,
	}
}
