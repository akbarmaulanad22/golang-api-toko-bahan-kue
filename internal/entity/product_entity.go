package entity

type Product struct {
	SKU        string `gorm:"column:sku;primaryKey"`
	Name       string `gorm:"column:name"`
	CategoryID uint   `gorm:"column:category_id"`
	CreatedAt  int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt  int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Category Category `gorm:"foreignKey:CategoryID;references:id"`
	Sizes    []Size   `gorm:"foreignKey:ProductSKU;references:sku"`
}
