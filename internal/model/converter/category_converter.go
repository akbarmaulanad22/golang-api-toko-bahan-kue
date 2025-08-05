package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func CategoryToResponse(category *entity.Category) *model.CategoryResponse {
	return &model.CategoryResponse{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}
