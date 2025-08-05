package entity

type SaleDetail struct {
	Qty         int   `gorm:"column:qty"`
	IsCancelled bool  `gorm:"column:is_cancelled;type:tinyint"`
	CancelledAt int64 `gorm:"column:cancelled_at"`

	SaleCode string `gorm:"column:sale_code;primaryKey"`
	SizeID   uint   `gorm:"column:size_id;primaryKey"`

	Size Size `gorm:"foreignKey:SizeID;references:ID"`
}
