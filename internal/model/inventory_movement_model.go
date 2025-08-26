package model

type BulkInventoryMovementResponse struct {
	ReferenceType string                      `json:"reference_type,omitempty"`
	ReferenceKey  string                      `json:"reference_key,omitempty"`
	Movements     []InventoryMovementResponse `json:"movements,omitempty"`
}

type InventoryMovementResponse struct {
	ID            uint   `json:"id,omitempty"`
	BranchName    string `json:"branch_name,omitempty"`
	ProductName   string `json:"product_name,omitempty"`
	SizeLabel     string `json:"size_label,omitempty"`
	ReferenceType string `json:"reference_type,omitempty"`
	ReferenceKey  string `json:"reference_key,omitempty"`
	ChangeQty     int    `json:"change_qty,omitempty"`
	CreatedAt     int64  `json:"created_at,omitempty"`
}

type BulkCreateInventoryMovementRequest struct {
	BranchID      uint                             `json:"-" validate:"required"`
	ReferenceType string                           `json:"reference_type" validate:"required"`
	ReferenceKey  string                           `json:"reference_key" validate:"required"`
	Movements     []CreateInventoryMovementRequest `json:"movements" validate:"required,dive"`
}

type CreateInventoryMovementRequest struct {
	SizeID    uint `json:"size_id" validate:"required"`
	ChangeQty int  `json:"change_qty" validate:"required"`
}

type SearchInventoryMovementRequest struct {
	BranchID *uint  `json:"-"`
	Type     string `json:"type"`
	Search   string `json:"search"` // bisa untuk search branch name/product name/size label/type/reference key
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
	Page     int    `json:"page" validate:"min=1"`
	Size     int    `json:"size" validate:"min=1,max=100"`
}
