package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func CapitalToResponse(e *entity.Capital) *model.CapitalResponse {
	return &model.CapitalResponse{
		ID:        e.ID,
		Type:      e.Type,
		Note:      e.Note,
		Amount:    e.Amount,
		BranchID:  e.BranchID,
		CreatedAt: e.CreatedAt,
	}
}
