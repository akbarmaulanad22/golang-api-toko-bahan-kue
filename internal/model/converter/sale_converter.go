package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SaleToResponse(sale *entity.Sale) *model.SaleResponse {

	return &model.SaleResponse{
		Code:         sale.Code,
		CustomerName: sale.CustomerName,
		Status:       sale.Status,
		CreatedAt:    sale.CreatedAt,
		BranchName:   sale.Branch.Name,
		TotalPrice:   sale.TotalPrice,
	}
}
