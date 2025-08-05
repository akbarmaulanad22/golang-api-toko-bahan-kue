package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SaleToResponse(sale *entity.Sale) *model.SaleResponse {
	return &model.SaleResponse{
		Code:         sale.Code,
		CustomerName: sale.CustomerName,
		Status:       model.StatusPayment(sale.Status),
		CashValue:    sale.CashValue,
		DebitValue:   sale.DebitValue,
		PaidAt:       sale.PaidAt,
		CreatedAt:    sale.CreatedAt,
		CancelledAt:  sale.CancelledAt,
		Branch:       *BranchToResponse(&sale.Branch),
	}
}
