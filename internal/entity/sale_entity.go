package entity

type Sale struct {
	Code         string        `gorm:"column:code;primaryKey"`
	CustomerName string        `gorm:"column:customer_name"`
	Status       string        `gorm:"column:status;default:COMPLETED"`
	BranchID     uint          `gorm:"column:branch_id"`
	CreatedAt    int64         `gorm:"column:created_at;autoCreateTime:milli"`
	Details      []SaleDetail  `gorm:"foreignKey:SaleCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Payments     []SalePayment `gorm:"foreignKey:SaleCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Debt         *Debt         `gorm:"foreignKey:ReferenceCode;references:Code;"`
	Branch       Branch        `gorm:"foreignKey:BranchID;references:ID;"`
	TotalPrice   float64       `gorm:"total_price"`
}
