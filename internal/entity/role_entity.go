package entity

type Role struct {
	ID        uint   `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
