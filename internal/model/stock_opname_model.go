package model

type StockOpnameResponse struct {
	ID          uint                        `json:"id"`
	BranchName  string                      `json:"branch_name"`
	Date        int64                       `json:"date"`
	Status      string                      `json:"status"`
	CreatedBy   string                      `json:"created_by"`
	VerifiedBy  string                      `json:"verified_by"`
	CreatedAt   int64                       `json:"created_at"`
	CompletedAt int64                       `json:"completed_at"`
	Details     []StockOpnameDetailResponse `json:"details"`
}

type CreateStockOpnameRequest struct {
	BranchID  uint                           `json:"-" validate:"required"`
	Details   []CreateStockOpnameDetailInput `json:"details" validate:"required,dive"`
	CreatedBy string                         `json:"-" validate:"required"`
}

type UpdateStockOpnameRequest struct {
	ID         uint                             `json:"-" validate:"required"`
	Details    []UpdateStockOpnameDetailRequest `json:"details" validate:"required,dive"`
	VerifiedBy string                           `json:"verified_by"`
}

type GetStockOpnameRequest struct {
	ID uint `json:"id" validate:"required"`
}

type DeleteStockOpnameRequest struct {
	ID uint `json:"-" validate:"required"`
}

type SearchStockOpnameRequest struct {
	BranchID  uint   `json:"branch_id" validate:"omitempty"`
	Status    string `json:"status" validate:"omitempty,oneof=draft completed cancelled"`
	CreatedBy string `json:"created_by" validate:"max=100"`
	DateFrom  int64  `json:"date_from" validate:"omitempty"`
	DateTo    int64  `json:"date_to" validate:"omitempty"`
	Page      int    `json:"page" validate:"min=1"`
	Size      int    `json:"size" validate:"min=1,max=100"`
}
