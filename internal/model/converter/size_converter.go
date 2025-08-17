package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SizeToResponse(size *entity.Size) *model.SizeResponse {
	var productResp *model.ProductResponse
	if size.Product != nil {
		productResp = ProductToResponse(size.Product)
	}

	return &model.SizeResponse{
		ID:        size.ID,
		Name:      size.Name,
		SellPrice: size.SellPrice,
		BuyPrice:  size.BuyPrice,
		CreatedAt: size.CreatedAt,
		UpdatedAt: size.UpdatedAt,
		Product:   productResp,
	}
}
