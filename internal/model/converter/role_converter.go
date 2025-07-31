package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func RoleToResponse(branch *entity.Role) *model.RoleResponse {
	return &model.RoleResponse{
		ID:        branch.ID,
		Name:      branch.Name,
		CreatedAt: branch.CreatedAt,
		UpdatedAt: branch.UpdatedAt,
	}
}
