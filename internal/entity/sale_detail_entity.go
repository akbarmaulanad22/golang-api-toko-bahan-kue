package entity

type SaleDetail struct {
	ID          uint    `gorm:"column:id;primaryKey"`
	Qty         int     `gorm:"column:qty"`
	SellPrice   float64 `gorm:"column:sell_price"`
	IsCancelled bool    `gorm:"column:is_cancelled;type:tinyint"`
	SaleCode    string  `gorm:"column:sale_code"`
	SizeID      uint    `gorm:"column:size_id"`

	Size *Size `gorm:"foreignKey:SizeID;references:ID"`
}
