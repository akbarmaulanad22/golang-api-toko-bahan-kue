package entity

type SaleDetail struct {
	ID                uint            `gorm:"column:id;primaryKey"`
	Qty               int             `gorm:"column:qty"`
	SellPrice         float64         `gorm:"column:sell_price"`
	IsCancelled       bool            `gorm:"column:is_cancelled;type:tinyint"`
	SaleCode          string          `gorm:"column:sale_code"`
	BranchInventoryID uint            `gorm:"column:branch_inventory_id"`
	BranchInventory   BranchInventory `gorm:"foreignKey:BranchInventoryID;references:ID"`
}
