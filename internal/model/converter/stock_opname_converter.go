package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func StockOpnameToResponse(stockOpname *entity.StockOpname) *model.StockOpnameResponse {
	if stockOpname == nil {
		return nil
	}

	// konversi details
	details := make([]model.StockOpnameDetailResponse, 0, len(stockOpname.Details))
	for _, d := range stockOpname.Details {
		details = append(details, *StockOpnameDetailToResponse(&d))
	}

	return &model.StockOpnameResponse{
		ID:          stockOpname.ID,
		BranchName:  stockOpname.Branch.Name, // akses dari relasi
		Date:        stockOpname.Date,
		Status:      stockOpname.Status,
		CreatedBy:   stockOpname.CreatedBy,
		VerifiedBy:  stockOpname.VerifiedBy,
		CreatedAt:   stockOpname.CreatedAt,
		CompletedAt: stockOpname.CompletedAt,
		Details:     details,
	}
}
