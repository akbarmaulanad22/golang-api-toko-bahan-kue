package entity

type DebtPayment struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	PaymentDate int64   `gorm:"column:payment_date"`
	Amount      float64 `gorm:"column:amount"`
	Note        string  `gorm:"column:note"`
	DebtID      uint    `gorm:"column:debt_id"`
	CreatedAt   int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
