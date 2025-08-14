package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func PurchaseToResponse(purchase *entity.Purchase) *model.PurchaseResponse {
	return &model.PurchaseResponse{
		Code:          purchase.Code,
		DistributorID: purchase.DistributorID,
		Status:        purchase.Status,
		BranchID:      purchase.BranchID,
		CreatedAt:     purchase.CreatedAt,
		Details:       []model.PurchaseDetailResponse{},
		Payments:      []model.PurchasePaymentResponse{},
	}
}
