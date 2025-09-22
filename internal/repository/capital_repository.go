package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CapitalRepository struct {
	Repository[entity.Capital]
	Log *logrus.Logger
}

func NewCapitalRepository(log *logrus.Logger) *CapitalRepository {
	return &CapitalRepository{
		Log: log,
	}
}

func (r *CapitalRepository) Search(db *gorm.DB, request *model.SearchCapitalRequest) ([]entity.Capital, int64, error) {
	var users []entity.Capital
	if err := db.Scopes(r.FilterCapital(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Capital{}).Scopes(r.FilterCapital(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *CapitalRepository) FilterCapital(request *model.SearchCapitalRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.BranchID != nil {
			tx = tx.Where("branch_id = ?", request.BranchID)
		}

		if note := request.Note; note != "" {
			note = "%" + note + "%"
			tx = tx.Where("note LIKE ?", note)
		}

		startAt := request.StartAt
		endAt := request.EndAt

		if startAt != 0 && endAt != 0 {
			tx = tx.Where("created_at BETWEEN ? AND ?", startAt, endAt)
		}

		return tx
	}
}

func (r *CapitalRepository) GetBalance(db *gorm.DB, branchID *uint) (float64, error) {
	var balance float64
	var err error

	// Base query
	sql := `
		SELECT COALESCE(SUM(
			CASE 
				WHEN type = 'IN' THEN amount
				WHEN type = 'OUT' THEN -amount
				ELSE 0
			END
		), 0) AS balance
		FROM capitals
	`

	if branchID != nil {
		// Filter by branch
		err = db.Raw(sql+` WHERE branch_id = ?`, *branchID).Scan(&balance).Error
	} else {
		// Semua cabang
		err = db.Raw(sql).Scan(&balance).Error
	}

	if err != nil {
		return 0, err
	}
	return balance, nil
}

// func (r *CapitalRepository) ConsolidateReport(db *gorm.DB, request *model.SearchConsolidateCapitalRequest) (*model.ConsolidatedCapitalResponse, error) {
// 	var reports []model.CapitalReportResponse
// 	var totalAll float64

// 	baseQuery := `
// 		SELECT b.id AS branch_id, b.name AS branch_name,
// 		       COALESCE(SUM(e.amount), 0) AS total_capitals
// 		FROM branches b
// 		LEFT JOIN capitals e ON e.branch_id = b.id
// 	`

// 	// Kondisi filter tanggal
// 	if request.StartAt != 0 && request.EndAt != 0 {
// 		baseQuery += " AND e.created_at BETWEEN ? AND ?"
// 	}

// 	baseQuery += " GROUP BY b.id, b.name ORDER BY b.name"

// 	// Query per cabang
// 	var err error
// 	if request.StartAt != 0 && request.EndAt != 0 {
// 		err = db.Raw(baseQuery, request.StartAt, request.EndAt).Scan(&reports).Error
// 	} else {
// 		err = db.Raw(baseQuery).Scan(&reports).Error
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Query total semua cabang
// 	totalQuery := `SELECT COALESCE(SUM(e.amount), 0) FROM capitals e`
// 	if request.StartAt != 0 && request.EndAt != 0 {
// 		totalQuery += " WHERE e.created_at BETWEEN ? AND ?"
// 		if err := db.Raw(totalQuery, request.StartAt, request.EndAt).Scan(&totalAll).Error; err != nil {
// 			return nil, err
// 		}
// 	} else {
// 		if err := db.Raw(totalQuery).Scan(&totalAll).Error; err != nil {
// 			return nil, err
// 		}
// 	}

// 	resp := &model.ConsolidatedCapitalResponse{
// 		Data:             reports,
// 		TotalAllBranches: totalAll,
// 	}

// 	return resp, nil

// }
