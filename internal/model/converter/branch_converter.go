package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func BranchToResponse(branch *entity.Branch) *model.BranchResponse {
	return &model.BranchResponse{
		ID:        branch.ID,
		Name:      branch.Name,
		Address:   branch.Address,
		CreatedAt: branch.CreatedAt,
		UpdatedAt: branch.UpdatedAt,
	}
}
