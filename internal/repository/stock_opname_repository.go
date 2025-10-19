package repository

import (
	"fmt"
	"strings"
	"tokobahankue/internal/entity"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StockOpnameRepository struct {
	Repository[entity.StockOpname]
	Log *logrus.Logger
}

func NewStockOpnameRepository(log *logrus.Logger) *StockOpnameRepository {
	return &StockOpnameRepository{
		Log: log,
	}
}

func (r *StockOpnameRepository) Search(db *gorm.DB, request *model.SearchStockOpnameRequest) ([]entity.StockOpname, int64, error) {
	var opnames []entity.StockOpname

	query := db.Model(&entity.StockOpname{}).
		Preload("Branch").
		Preload("Details").
		Scopes(r.FilterStockOpname(request))

	if err := query.Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&opnames).Error; err != nil {
		return nil, 0, err
	}

	// hitung total data
	var total int64
	if err := db.Model(&entity.StockOpname{}).Scopes(r.FilterStockOpname(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return opnames, total, nil
}

func (r *StockOpnameRepository) FilterStockOpname(request *model.SearchStockOpnameRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.BranchID != 0 {
			tx = tx.Where("branch_id = ?", request.BranchID)
		}

		if request.Status != "" {
			tx = tx.Where("status = ?", request.Status)
		}

		if request.CreatedBy != "" {
			tx = tx.Where("created_by LIKE ?", "%"+request.CreatedBy+"%")
		}

		// if request.DateFrom != 0 && request.DateTo != 0 {
		// 	tx = tx.Where("date BETWEEN ? AND ?", request.DateFrom, request.DateTo)
		// }

		return tx
	}
}

func (r *StockOpnameRepository) Create(db *gorm.DB, opname *entity.StockOpname) error {
	return db.Session(&gorm.Session{FullSaveAssociations: true}).Create(opname).Error
}

func (r *StockOpnameRepository) FindByIdWithDetails(db *gorm.DB, opname *entity.StockOpname, id uint) error {
	return db.Preload("Details.BranchInventory.Size").Preload("Branch").First(opname, id).Error
}

func (r *StockOpnameRepository) BulkUpdateDetails(tx *gorm.DB, opnameID int64, details []entity.StockOpnameDetail) error {
	if len(details) == 0 {
		return nil
	}

	var ids []int64
	casePhysical := "CASE branch_inventory_id"
	caseNotes := "CASE branch_inventory_id"

	for _, d := range details {
		ids = append(ids, int64(d.BranchInventoryID))
		casePhysical += fmt.Sprintf(" WHEN %d THEN %d", d.BranchInventoryID, d.PhysicalQty)
		caseNotes += fmt.Sprintf(" WHEN %d THEN '%s'", d.BranchInventoryID, strings.ReplaceAll(d.Notes, "'", "''"))
	}

	casePhysical += " END"
	caseNotes += " END"

	query := fmt.Sprintf(`
		UPDATE stock_opname_detail
		SET 
			physical_qty = %s,
			notes = %s
		WHERE stock_opname_id = ? AND branch_inventory_id IN (%s)
	`, casePhysical, caseNotes, helper.Int64Join(ids, ","))

	return tx.Exec(query, opnameID).Error
}
