package entity

type Purchase struct {
	Code          string            `gorm:"column:code;primaryKey"`
	SalesName     string            `gorm:"column:sales_name"`
	Status        string            `gorm:"column:status;default:COMPLETED"`
	BranchID      uint              `gorm:"column:branch_id"`
	DistributorID uint              `gorm:"column:distributor_id"`
	CreatedAt     int64             `gorm:"column:created_at;autoCreateTime:milli"`
	Details       []PurchaseDetail  `gorm:"foreignKey:PurchaseCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Payments      []PurchasePayment `gorm:"foreignKey:PurchaseCode;references:Code;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Debt          *Debt             `gorm:"foreignKey:ReferenceCode;references:Code;"`
	Branch        Branch            `gorm:"foreignKey:BranchID;references:ID;"`
	Distributor   Distributor       `gorm:"foreignKey:DistributorID;references:ID;"`
	TotalPrice    float64           `gorm:"total_price"`
}
