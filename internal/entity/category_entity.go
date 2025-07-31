package entity

type Category struct {
	Slug      string `gorm:"column:slug;primaryKey"`
	Name      string `gorm:"column:name"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
