package converter

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"
)

func SaleToResponse(sale *entity.Sale) *model.SaleResponse {
	var debtResp *model.DebtResponse
	if sale.Debt != nil {
		debtResp = DebtToResponse(sale.Debt)
	}

	payments := make([]model.SalePaymentResponse, len(sale.Payments))
	for i, payment := range sale.Payments {
		payments[i] = *SalePaymentToResponse(&payment)
	}

	details := make([]model.SaleDetailResponse, len(sale.Details))
	for i, detail := range sale.Details {
		details[i] = *SaleDetailToResponse(&detail)
	}

	return &model.SaleResponse{
		Code:         sale.Code,
		CustomerName: sale.CustomerName,
		Status:       sale.Status,
		CreatedAt:    sale.CreatedAt,
		BranchID:     sale.BranchID,
		Details:      details,
		Payments:     payments,
		Debt:         debtResp,
	}
}
