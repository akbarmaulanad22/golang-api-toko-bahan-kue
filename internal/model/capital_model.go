package model

type CapitalResponse struct {
	ID        uint    `json:"id,omitempty"`
	Type      string  `json:"type,omitempty"`
	Note      string  `json:"note,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	BranchID  *uint   `json:"branch_id,omitempty"`
	CreatedAt int64   `json:"created_at,omitempty"`
}

type CreateCapitalRequest struct {
	Type     string  `json:"type" validate:"required,oneof=IN OUT"`
	Note     string  `json:"note" validate:"max=255"`
	Amount   float64 `json:"amount" validate:"required"`
	BranchID *uint   `json:"branch_id"`
}

type UpdateCapitalRequest struct {
	ID     uint    `json:"-" validate:"required"`
	Type   string  `json:"type" validate:"required,oneof=IN OUT"`
	Note   string  `json:"note" validate:"max=255"`
	Amount float64 `json:"amount" validate:"required"`
}

type SearchCapitalRequest struct {
	BranchID *uint  `json:"branch_id"`
	Type     string `json:"type"`
	Note     string `json:"note" validate:"max=255"`
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}

type DeleteCapitalRequest struct {
	ID uint `json:"-" validate:"required"`
}

// type SearchConsolidateCapitalRequest struct {
// 	StartAt int64 `json:"start_at"`
// 	EndAt   int64 `json:"end_at"`
// }

// type CapitalReportResponse struct {
// 	BranchID      uint    `json:"branch_id"`
// 	BranchName    string  `json:"branch_name"`
// 	TotalCapitals float64 `json:"total_expenses"`
// }

// type ConsolidatedCapitalResponse struct {
// 	Data             []CapitalReportResponse `json:"data"`
// 	TotalAllBranches float64                 `json:"total_all_branches"`
// }
