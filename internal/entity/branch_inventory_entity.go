package entity

type BranchInventory struct {
	ID        uint  `gorm:"column:id;primaryKey"`
	Stock     int   `gorm:"column:stock"`
	BranchID  uint  `gorm:"column:branch_id"`
	SizeID    uint  `gorm:"column:size_id"`
	CreatedAt int64 `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64 `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}

func (BranchInventory) TableName() string {
	return "branch_inventory"
}
