package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func CategoryToResponse(category *entity.Category) *model.CategoryResponse {
	return &model.CategoryResponse{
		Slug:      category.Slug,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
