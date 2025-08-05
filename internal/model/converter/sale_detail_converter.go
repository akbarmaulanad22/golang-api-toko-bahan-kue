package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SaleDetailToResponse(saleDetail *entity.SaleDetail) *model.SaleDetailResponse {
	return &model.SaleDetailResponse{
		SaleCode:    saleDetail.SaleCode,
		SizeID:      saleDetail.SizeID,
		Qty:         saleDetail.Qty,
		IsCancelled: saleDetail.IsCancelled,
		Size:        *SizeToResponse(&saleDetail.Size),
	}
}
