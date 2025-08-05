package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func ProductToResponse(product *entity.Product) *model.ProductResponse {
	sizes := make([]model.SizeResponse, len(product.Sizes))
	for i, size := range product.Sizes {
		sizes[i] = *SizeToResponse(&size)
	}

	return &model.ProductResponse{
		SKU:       product.SKU,
		Name:      product.Name,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
		Category:  *CategoryToResponse(&product.Category),
		Sizes:     sizes,
	}
}
