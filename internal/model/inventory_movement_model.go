package model

type BulkInventoryMovementResponse struct {
	ReferenceType string                      `json:"reference_type,omitempty"`
	ReferenceKey  string                      `json:"reference_key,omitempty"`
	Movements     []InventoryMovementResponse `json:"movements,omitempty"`
}

type InventoryMovementResponse struct {
	ID        uint  `json:"id,omitempty"`
	ChangeQty int   `json:"change_qty,omitempty"`
	CreatedAt int64 `json:"created_at,omitempty"`
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
