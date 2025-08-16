package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SaleDetailToResponse(saleDetail *entity.SaleDetail) *model.SaleDetailResponse {
	return &model.SaleDetailResponse{
		SizeID:      saleDetail.SizeID,
		Qty:         saleDetail.Qty,
		SellPrice:   saleDetail.SellPrice,
		IsCancelled: saleDetail.IsCancelled,
		Size:        SizeToResponse(saleDetail.Size),
	}
}
