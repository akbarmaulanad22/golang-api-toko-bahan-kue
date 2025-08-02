package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SizeToResponse(size *entity.Size) *model.SizeResponse {
	return &model.SizeResponse{
		ID:        size.ID,
		Name:      size.Name,
		SellPrice: size.SellPrice,
		BuyPrice:  size.BuyPrice,
		CreatedAt: size.CreatedAt,
		UpdatedAt: size.UpdatedAt,
		// Product:   *ProductToResponse(&size.Product),
	}
}
