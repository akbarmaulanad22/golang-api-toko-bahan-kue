package entity

type Expense struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	Description string  `gorm:"column:description"`
	Amount      float64 `gorm:"column:amount"`
	BranchID    uint    `gorm:"column:branch_id"`
	CreatedAt   int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
