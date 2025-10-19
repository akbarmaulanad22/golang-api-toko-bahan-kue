package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func StockOpnameDetailToResponse(detail *entity.StockOpnameDetail) *model.StockOpnameDetailResponse {
	if detail == nil {
		return nil
	}

	return &model.StockOpnameDetailResponse{
		ID:                detail.ID,
		StockOpnameID:     detail.StockOpnameID,
		BranchInventoryID: detail.BranchInventoryID,
		SystemQty:         detail.SystemQty,
		PhysicalQty:       detail.PhysicalQty,
		Difference:        detail.Difference,
		Notes:             detail.Notes,
	}
}
