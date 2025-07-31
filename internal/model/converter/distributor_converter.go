package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func DistributorToResponse(branch *entity.Distributor) *model.DistributorResponse {
	return &model.DistributorResponse{
		ID:        branch.ID,
		Name:      branch.Name,
		Address:   branch.Address,
		CreatedAt: branch.CreatedAt,
		UpdatedAt: branch.UpdatedAt,
	}
}
