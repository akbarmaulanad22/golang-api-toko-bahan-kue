package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func PurchaseDetailToResponse(purchaseDetail *entity.PurchaseDetail) *model.PurchaseDetailResponse {
	return &model.PurchaseDetailResponse{
		SizeID:      purchaseDetail.SizeID,
		Qty:         purchaseDetail.Qty,
		BuyPrice:    purchaseDetail.BuyPrice,
		IsCancelled: purchaseDetail.IsCancelled,
	}
}
