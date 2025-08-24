package entity

type InventoryMovement struct {
	ID                uint   `gorm:"column:id;primaryKey"`
	ChangeQty         uint   `gorm:"column:change_qty"`
	BranchInventoryID uint   `gorm:"column:branch_inventory_id"`
	ReferenceType     string `gorm:"column:reference_type"`
	ReferenceKey      string `gorm:"column:reference_key"`
	CreatedAt         int64  `gorm:"column:created_at;autoCreateTime:milli"`
}
