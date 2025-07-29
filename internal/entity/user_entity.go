package entity

type User struct {
	Username  string `gorm:"column:username;primaryKey"`
	Password  string `gorm:"column:password"`
	Token     string `gorm:"column:token"`
	Name      string `gorm:"column:name"`
	Address   string `gorm:"column:address"`
	CreatedAt int64  `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`
}
