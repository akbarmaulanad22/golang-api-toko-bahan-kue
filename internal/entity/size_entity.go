package entity

type Size struct {
	ID         uint   `gorm:"column:id;primaryKey"`
	Name       string `gorm:"column:name"`
	SellPrice  uint   `gorm:"column:sell_price"`
	BuyPrice   uint   `gorm:"column:buy_price"`
	ProductSKU string `gorm:"column:product_sku"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Product Product `gorm:"foreignKey:ProductSKU;references:sku"`
}
