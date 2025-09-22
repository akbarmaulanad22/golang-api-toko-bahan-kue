package entity

type PurchaseDetail struct {
	Qty         int     `gorm:"column:qty"`
	BuyPrice    float64 `gorm:"column:buy_price"`
	IsCancelled bool    `gorm:"column:is_cancelled;type:tinyint"`

	PurchaseCode string `gorm:"column:purchase_code;primaryKey"`
	SizeID       uint   `gorm:"column:size_id;primaryKey"`

	Size *Size `gorm:"foreignKey:SizeID;references:ID"`
}
