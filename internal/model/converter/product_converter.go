package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
	return &model.ProductResponse{
		SKU:       product.SKU,
		Name:      product.Name,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		Category:  *CategoryToResponse(&product.Category),
	}
}
