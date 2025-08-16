package entity

type Debt struct {
	ID            uint    `gorm:"column:id;primaryKey"`
	ReferenceType string  `gorm:"column:reference_type"`
	TotalAmount   float64 `gorm:"column:total_amount"`
	PaidAmount    float64 `gorm:"column:paid_amount"`
	DueDate       int64   `gorm:"column:due_date"`
	Status        string  `gorm:"column:status"`
	ReferenceCode string  `gorm:"column:reference_code"`
	CreatedAt     int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt     int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Payments []DebtPayment `gorm:"foreignKey:DebtID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
