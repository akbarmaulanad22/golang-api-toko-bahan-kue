package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ExpenseRepository struct {
	Repository[entity.Expense]
	Log *logrus.Logger
}

func NewExpenseRepository(log *logrus.Logger) *ExpenseRepository {
	return &ExpenseRepository{
		Log: log,
	}
}

func (r *ExpenseRepository) Search(db *gorm.DB, request *model.SearchExpenseRequest) ([]entity.Expense, int64, error) {
	var users []entity.Expense
	if err := db.Scopes(r.FilterExpense(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Expense{}).Scopes(r.FilterExpense(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *ExpenseRepository) FilterExpense(request *model.SearchExpenseRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if request.BranchID != nil {
			tx = tx.Where("branch_id = ?", request.BranchID)
		}

		if description := request.Description; description != "" {
			description = "%" + description + "%"
			tx = tx.Where("description LIKE ?", description)
		}

		startAt := request.StartAt
		endAt := request.EndAt

		if startAt != 0 && endAt != 0 {
			tx = tx.Where("created_at BETWEEN ? AND ?", startAt, endAt)
		}

		return tx
	}
}

func (r *ExpenseRepository) ConsolidateReport(db *gorm.DB, request *model.SearchConsolidateExpenseRequest) (*model.ConsolidatedExpenseResponse, error) {
	var reports []model.ExpenseReportResponse
	var totalAll float64

	baseQuery := `
		SELECT b.id AS branch_id, b.name AS branch_name,
		       COALESCE(SUM(e.amount), 0) AS total_expenses
		FROM branches b
		LEFT JOIN expenses e ON e.branch_id = b.id
	`

	// Kondisi filter tanggal
	if request.StartAt != 0 && request.EndAt != 0 {
		baseQuery += " AND e.created_at BETWEEN ? AND ?"
	}

	baseQuery += " GROUP BY b.id, b.name ORDER BY b.name"

	// Query per cabang
	var err error
	if request.StartAt != 0 && request.EndAt != 0 {
		err = db.Raw(baseQuery, request.StartAt, request.EndAt).Scan(&reports).Error
	} else {
		err = db.Raw(baseQuery).Scan(&reports).Error
	}
	if err != nil {
		return nil, err
	}

	// Query total semua cabang
	totalQuery := `SELECT COALESCE(SUM(e.amount), 0) FROM expenses e`
	if request.StartAt != 0 && request.EndAt != 0 {
		totalQuery += " WHERE e.created_at BETWEEN ? AND ?"
		if err := db.Raw(totalQuery, request.StartAt, request.EndAt).Scan(&totalAll).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Raw(totalQuery).Scan(&totalAll).Error; err != nil {
			return nil, err
		}
	}

	resp := &model.ConsolidatedExpenseResponse{
		Data:             reports,
		TotalAllBranches: totalAll,
	}

	return resp, nil

}
