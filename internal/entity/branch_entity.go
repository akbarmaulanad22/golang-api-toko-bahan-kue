package entity

type Branch struct {
	ID        uint   `gorm:"column:id;primaryKey"`
	Name      string `gorm:"column:name"`
	Address   string `gorm:"column:address"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Sales []Sale `gorm:"foreignKey:BranchID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
