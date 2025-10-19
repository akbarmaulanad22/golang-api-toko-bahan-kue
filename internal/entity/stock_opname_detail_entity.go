package entity

type StockOpnameDetail struct {
	ID                uint            `gorm:"column:id;primaryKey"`
	StockOpnameID     uint            `gorm:"column:stock_opname_id"`
	BranchInventoryID uint            `gorm:"column:branch_inventory_id"`
	SystemQty         int64           `gorm:"column:system_qty"`
	PhysicalQty       int64           `gorm:"column:physical_qty"`
	Difference        int64           `gorm:"column:difference"`
	Notes             string          `gorm:"column:notes"`
	BranchInventory   BranchInventory `gorm:"foreignKey:BranchInventoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (*StockOpnameDetail) TableName() string {
	return "stock_opname_detail"
}
