package entity

type CashBankTransaction struct {
	ID              uint    `gorm:"column:id;primaryKey"`
	TransactionDate int64   `gorm:"column:transaction_date;autoCreateTime:milli"`
	Type            string  `gorm:"column:type"`
	Source          string  `gorm:"column:source"`
	Amount          float64 `gorm:"column:amount"`
	Description     string  `gorm:"column:description"`
	ReferenceKey    string  `gorm:"column:reference_key"`
	BranchID        *uint   `gorm:"column:branch_id"`
	CreatedAt       int64   `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt       int64   `gorm:"column:updated_at;autoCreateTime:milli;autoUpdateTime:milli"`

	Branch *Branch `gorm:"foreignKey:BranchID;references:ID;"`
}
