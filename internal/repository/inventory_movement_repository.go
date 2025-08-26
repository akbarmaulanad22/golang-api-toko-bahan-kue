package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InventoryMovementRepository struct {
	Repository[entity.InventoryMovement]
	Log *logrus.Logger
}

func NewInventoryMovementRepository(log *logrus.Logger) *InventoryMovementRepository {
	return &InventoryMovementRepository{
		Log: log,
	}
}

func (r *InventoryMovementRepository) Search(db *gorm.DB, request *model.SearchInventoryMovementRequest) ([]model.InventoryMovementResponse, int64, error) {
	var movements []model.InventoryMovementResponse

	// Query utama dengan join
	baseQuery := db.Table("inventory_movements im").
		Select(`
			im.id,
			b.name AS branch_name,
			p.name AS product_name,
			s.name AS size_label,
			im.reference_type,
			im.reference_key,
			im.change_qty,
			im.created_at
		`).
		Joins("JOIN branch_inventory bi ON im.branch_inventory_id = bi.id").
		Joins("JOIN branches b ON bi.branch_id = b.id").
		Joins("JOIN sizes s ON bi.size_id = s.id").
		Joins("JOIN products p ON s.product_sku = p.sku").
		Scopes(r.FilterInventoryMovements(request))

	// Ambil data
	if err := baseQuery.
		Offset((request.Page - 1) * request.Size).
		Limit(request.Size).
		Order("im.created_at DESC").
		Scan(&movements).Error; err != nil {
		return nil, 0, err
	}

	// Hitung total
	var total int64
	if err := db.Table("inventory_movements im").
		Joins("JOIN branch_inventory bi ON im.branch_inventory_id = bi.id").
		Joins("JOIN branches b ON bi.branch_id = b.id").
		Joins("JOIN sizes s ON bi.size_id = s.id").
		Joins("JOIN products p ON s.product_sku = p.sku").
		Scopes(r.FilterInventoryMovements(request)).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return movements, total, nil
}

func (r *InventoryMovementRepository) FilterInventoryMovements(request *model.SearchInventoryMovementRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		// Filter Branch
		if request.BranchID != nil {
			tx = tx.Where("bi.branch_id = ?", request.BranchID)
		}

		// Filter Type
		if request.Type != "" {
			tx = tx.Where("im.reference_type = ?", request.Type)
		}

		// Filter Date
		if request.StartAt > 0 && request.EndAt > 0 {
			tx = tx.Where("im.created_at BETWEEN ? AND ?", request.StartAt, request.EndAt)
		}

		// Search
		if request.Search != "" {
			searchLike := "%" + request.Search + "%"
			tx = tx.Where(`
				b.name LIKE ? OR 
				p.name LIKE ? OR 
				s.label LIKE ? OR 
				im.reference_type LIKE ? OR 
				im.reference_key LIKE ?`,
				searchLike, searchLike, searchLike, searchLike, searchLike)
		}

		return tx
	}
}

func (r *InventoryMovementRepository) SummaryByBranch(db *gorm.DB, request *model.SearchInventoryMovementRequest) (*model.InventoryMovementSummaryResponse, error) {
	var summaries []model.InventoryMovementBranchSummary

	// per branch
	if err := db.Table("inventory_movements im").
		Select(`
			b.id AS branch_id,
			b.name AS branch_name,
			COALESCE(SUM(CASE WHEN im.change_qty > 0 THEN im.change_qty ELSE 0 END),0) AS total_in,
			COALESCE(SUM(CASE WHEN im.change_qty < 0 THEN ABS(im.change_qty) ELSE 0 END),0) AS total_out
		`).
		Joins("JOIN branch_inventory bi ON im.branch_inventory_id = bi.id").
		Joins("JOIN branches b ON bi.branch_id = b.id").
		Joins("JOIN sizes s ON bi.size_id = s.id").
		Joins("JOIN products p ON s.product_sku = p.sku").
		Scopes(r.FilterInventoryMovements(request)).
		Group("b.id, b.name").
		Scan(&summaries).Error; err != nil {
		return nil, err
	}

	// total all branches
	var total model.InventoryMovementSummaryAll
	if err := db.Table("inventory_movements im").
		Select(`
			COALESCE(SUM(CASE WHEN im.change_qty > 0 THEN im.change_qty ELSE 0 END),0) AS total_in,
			COALESCE(SUM(CASE WHEN im.change_qty < 0 THEN ABS(im.change_qty) ELSE 0 END),0) AS total_out
		`).
		Joins("JOIN branch_inventory bi ON im.branch_inventory_id = bi.id").
		Joins("JOIN branches b ON bi.branch_id = b.id").
		Joins("JOIN sizes s ON bi.size_id = s.id").
		Joins("JOIN products p ON s.product_sku = p.sku").
		Scopes(r.FilterInventoryMovements(request)).
		Scan(&total).Error; err != nil {
		return nil, err
	}

	return &model.InventoryMovementSummaryResponse{
		Data:             summaries,
		TotalAllBranches: total,
	}, nil
}
