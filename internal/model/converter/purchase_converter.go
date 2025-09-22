package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func PurchaseToResponse(purchase *entity.Purchase) *model.PurchaseResponse {

	return &model.PurchaseResponse{
		Code:            purchase.Code,
		SalesName:       purchase.SalesName,
		Status:          purchase.Status,
		CreatedAt:       purchase.CreatedAt,
		BranchName:      purchase.Branch.Name,
		DistributorName: purchase.Distributor.Name,
		TotalPrice:      purchase.TotalPrice,
	}
}
