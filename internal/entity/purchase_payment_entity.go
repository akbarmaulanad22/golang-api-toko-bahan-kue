package entity

type PurchasePayment struct {
	ID            uint    `gorm:"column:id;primaryKey"`
	PurchaseCode  string  `gorm:"column:sale_code"`
	PaymentMethod string  `gorm:"column:payment_method"`
	Amount        float64 `gorm:"column:amount"`
	Note          string  `gorm:"column:note"`
	CreatedAt     int64   `gorm:"column:created_at;autoCreateTime:milli"`
}
