package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DebtRepository struct {
	Repository[entity.Debt]
	Log *logrus.Logger
}

func NewDebtRepository(log *logrus.Logger) *DebtRepository {
	return &DebtRepository{
		Log: log,
	}
}

func (r *DebtRepository) SearchRaw(db *gorm.DB, request *model.SearchDebtRequest) ([]model.DebtResponse, int64, error) {
	var results []model.DebtResponse
	var total int64

	// Base query
	query := `
		SELECT 
			d.id,
			d.reference_type,
			d.reference_code,
			d.total_amount,
			d.paid_amount,
			d.due_date,
			d.status,
			CASE 
				WHEN d.reference_type = 'SALE' THEN s.customer_name
				WHEN d.reference_type = 'PURCHASE' THEN p.sales_name
				ELSE ''
			END AS related
		FROM debts d
		LEFT JOIN sales s ON d.reference_type = 'SALE' AND d.reference_code = s.code
		LEFT JOIN purchases p ON d.reference_type = 'PURCHASE' AND d.reference_code = p.code
		WHERE 1=1
	`

	// Params slice
	var params []interface{}

	// ReferenceType filter
	if request.ReferenceType != "" {
		query += " AND d.reference_type = ?"
		params = append(params, request.ReferenceType)
	}

	// ReferenceCode filter
	if request.ReferenceCode != "" {
		query += " AND d.reference_code LIKE ?"
		params = append(params, "%"+request.ReferenceCode+"%")
	}

	// Date filter
	if request.StartAt != 0 && request.EndAt != 0 {
		query += " AND ((d.due_date BETWEEN ? AND ?) OR (d.created_at BETWEEN ? AND ?))"
		params = append(params, request.StartAt, request.EndAt, request.StartAt, request.EndAt)
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS count_table"
	if err := db.Raw(countQuery, params...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	query += " ORDER BY d.id DESC LIMIT ? OFFSET ?"
	params = append(params, request.Size, (request.Page-1)*request.Size)

	// Main query
	if err := db.Raw(query, params...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// // Payments load per debt
	// for i := range results {
	// 	var payments []model.DebtPaymentResponse
	// 	if err := db.Raw(`
	// 		SELECT
	// 			dp.id,
	// 			dp.debt_id,
	// 			dp.amount,
	// 			dp.paid_at
	// 		FROM debt_payments dp
	// 		WHERE dp.debt_id = ?
	// 		ORDER BY dp.paid_at ASC
	// 	`, results[i].ID).Scan(&payments).Error; err != nil {
	// 		return nil, 0, err
	// 	}
	// 	results[i].Payments = payments
	// }

	return results, total, nil
}

// func (r *DebtRepository) Search(db *gorm.DB, request *model.SearchDebtRequest) ([]entity.Debt, int64, error) {
// 	var users []entity.Debt
// 	if err := db.Scopes(r.FilterDebt(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&users).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	var total int64 = 0
// 	if err := db.Model(&entity.Debt{}).Scopes(r.FilterDebt(request)).Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	return users, total, nil
// }

// func (r *DebtRepository) FilterDebt(request *model.SearchDebtRequest) func(tx *gorm.DB) *gorm.DB {
// 	return func(tx *gorm.DB) *gorm.DB {
// 		if referenceType := request.ReferenceType; referenceType != "" {
// 			referenceType = "%" + referenceType + "%"
// 			tx = tx.Where("reference_type = ?", referenceType)
// 		}

// 		if referenceCode := request.ReferenceCode; referenceCode != "" {
// 			referenceCode = "%" + referenceCode + "%"
// 			tx = tx.Where("reference_code LIKE ?", referenceCode)
// 		}

// 		startAt := request.StartAt
// 		endAt := request.EndAt

// 		if startAt != 0 && endAt != 0 {
// 			tx = tx.Where("(due_date BETWEEN ? AND ?) OR (created_at BETWEEN ? AND ?)", startAt, endAt, startAt, endAt)
// 		}

// 		return tx
// 	}
// }
