package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func PurchaseDetailToResponse(purchaseDetail *entity.PurchaseDetail) *model.PurchaseDetailResponse {
	return &model.PurchaseDetailResponse{
		PurchaseCode: purchaseDetail.PurchaseCode,
		SizeID:       purchaseDetail.SizeID,
		Qty:          purchaseDetail.Qty,
		IsCancelled:  purchaseDetail.IsCancelled,
		Size:         *SizeToResponse(&purchaseDetail.Size),
	}
}
