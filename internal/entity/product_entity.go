package entity

type Product struct {
	SKU          string `gorm:"column:sku;primaryKey"`
	Name         string `gorm:"column:name"`
	CategorySlug string `gorm:"column:category_slug"`
	CreatedAt    int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt    int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Category Category `gorm:"foreignKey:CategorySlug;references:slug"`
}
