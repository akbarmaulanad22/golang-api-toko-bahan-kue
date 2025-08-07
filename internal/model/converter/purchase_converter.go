package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func PurchaseToResponse(purchase *entity.Purchase) *model.PurchaseResponse {
	return &model.PurchaseResponse{
		Code:        purchase.Code,
		SalesName:   purchase.SalesName,
		Status:      model.StatusPayment(purchase.Status),
		CashValue:   purchase.CashValue,
		DebitValue:  purchase.DebitValue,
		PaidAt:      purchase.PaidAt,
		CreatedAt:   purchase.CreatedAt,
		CancelledAt: purchase.CancelledAt,
		Branch:      *BranchToResponse(&purchase.Branch),
		Distributor: *DistributorToResponse(&purchase.Distributor),
	}
}
