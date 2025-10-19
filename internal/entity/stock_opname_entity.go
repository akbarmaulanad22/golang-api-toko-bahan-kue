package entity

type StockOpname struct {
	ID          uint   `gorm:"column:id;primaryKey"`
	BranchID    uint   `gorm:"column:branch_id"`
	Date        int64  `gorm:"column:date;autoCreateTime:milli"`
	Status      string `gorm:"column:status;default:draft"`
	CreatedBy   string `gorm:"column:created_by"`
	VerifiedBy  string `gorm:"column:verified_by"`
	CreatedAt   int64  `gorm:"column:created_at;autoCreateTime:milli"`
	CompletedAt int64  `gorm:"column:completed_at"`

	Branch  Branch              `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Details []StockOpnameDetail `gorm:"foreignKey:StockOpnameID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (*StockOpname) TableName() string {
	return "stock_opname"
}
