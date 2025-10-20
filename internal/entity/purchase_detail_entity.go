package entity

type PurchaseDetail struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	Qty         int     `gorm:"column:qty"`
	BuyPrice    float64 `gorm:"column:buy_price"`
	IsCancelled bool    `gorm:"column:is_cancelled;type:tinyint"`

	PurchaseCode string `gorm:"column:purchase_code"`
	SizeID       uint   `gorm:"column:size_id"`

	Size *Size `gorm:"foreignKey:SizeID;references:ID"`
}
