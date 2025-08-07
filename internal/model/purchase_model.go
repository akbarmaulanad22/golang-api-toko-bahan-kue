package model

type PurchaseResponse struct {
	Code        string        `json:"code,omitempty"`
	SalesName   string        `json:"sales_name,omitempty"`
	Status      StatusPayment `json:"status,omitempty"`
	CashValue   int           `json:"cash_value,omitempty"`
	DebitValue  int           `json:"debit_value,omitempty"`
	PaidAt      int64         `json:"paid_at,omitempty"`
	CreatedAt   int64         `json:"created_at,omitempty"`
	CancelledAt int64         `json:"cancelled_at,omitempty"`

	Branch          BranchResponse           `json:"branch"`
	Distributor     DistributorResponse      `json:"distributor"`
	PurchaseDetails []PurchaseDetailResponse `json:"purchase_details"`
}

type CreatePurchaseRequest struct {
	SalesName       string                        `json:"sales_name" validate:"required,max=100"`
	Status          StatusPayment                 `json:"status" validate:"required,oneof=PENDING COMPLETED CANCELLED"`
	CashValue       int                           `json:"cash_value" validate:"required"`
	DebitValue      int                           `json:"debit_value" validate:"required"`
	PaidAt          int64                         `json:"-"`
	BranchID        uint                          `json:"-" validate:"required"`
	DistributorID   uint                          `json:"distributor_id" validate:"required"`
	PurchaseDetails []CreatePurchaseDetailRequest `json:"purchase_details" validate:"required"`
}

type SearchPurchaseRequest struct {
	Code      string        `json:"code" validate:"max=100"`
	SalesName string        `json:"sales_name" validate:"max=100"`
	Status    StatusPayment `json:"status"`
	StartAt   int64         `json:"start_at"`
	EndAt     int64         `json:"end_at"`
	Page      int           `json:"page" validate:"min=1"`
	Size      int           `json:"size" validate:"min=1,max=100"`
}

type GetPurchaseRequest struct {
	Code string `json:"-" validate:"required,max=100"`
}

type UpdatePurchaseRequest struct {
	Code   string        `json:"-" validate:"required,max=100"`
	Status StatusPayment `json:"status" validate:"required,oneof=PENDING COMPLETED CANCELLED"`
}
