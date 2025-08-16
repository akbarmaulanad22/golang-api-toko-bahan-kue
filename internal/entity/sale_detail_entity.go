package entity

type SaleDetail struct {
	Qty         int     `gorm:"column:qty"`
	SellPrice   float64 `gorm:"column:sell_price"`
	IsCancelled bool    `gorm:"column:is_cancelled;type:tinyint"`

	SaleCode string `gorm:"column:sale_code;primaryKey"`
	SizeID   uint   `gorm:"column:size_id;primaryKey"`

	Size *Size `gorm:"foreignKey:SizeID;references:ID"`
}
