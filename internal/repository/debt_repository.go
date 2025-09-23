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
	results := make([]model.DebtResponse, 0)
	var total int64

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
			END AS related,
			COALESCE(bs.name, bp.name, '') AS branch_name,
			d.created_at
		FROM debts d
		LEFT JOIN sales s ON d.reference_type = 'SALE' AND d.reference_code = s.code
		LEFT JOIN purchases p ON d.reference_type = 'PURCHASE' AND d.reference_code = p.code
		LEFT JOIN branches bs ON s.branch_id = bs.id
		LEFT JOIN branches bp ON p.branch_id = bp.id
		WHERE 1=1
	`

	var params []interface{}

	// BranchID filter
	if request.BranchID != nil {
		query += " AND (bp.id = ? OR bs.id = ?)"
		params = append(params, request.BranchID, request.BranchID)
	}

	// Status filter
	if request.Status != "" {
		query += " AND d.status = ?"
		params = append(params, request.Status)
	}

	// ReferenceType filter
	if request.ReferenceType != "" {
		query += " AND d.reference_type = ?"
		params = append(params, request.ReferenceType)
	}

	// ReferenceCode filter
	if request.Search != "" {
		query += " AND d.reference_code LIKE ? OR s.customer_name LIKE ? OR p.sales_name LIKE ?"
		params = append(params, "%"+request.Search+"%", "%"+request.Search+"%", "%"+request.Search+"%")
	}

	// Date filter
	if request.StartAt != 0 && request.EndAt != 0 {
		query += " AND ((d.due_date BETWEEN ? AND ?) OR (d.created_at BETWEEN ? AND ?))"
		params = append(params, request.StartAt, request.EndAt, request.StartAt, request.EndAt)
	}

	// Count
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

	return results, total, nil
}

func (r *DebtRepository) FindDetailById(db *gorm.DB, request *model.GetDebtRequest) (*model.DebtDetailResponse, error) {
	var debt model.DebtDetailResponse

	// Ambil data utama hutang
	if err := db.Raw(`
		SELECT id, reference_type, reference_code, total_amount, paid_amount, due_date, status, created_at
		FROM debts
		WHERE id = ?
	`, request.ID).Scan(&debt).Error; err != nil {
		return nil, err
	}

	// Ambil payments
	var payments []model.DebtPaymentResponse
	if err := db.Raw(`
		SELECT id, payment_date, amount, note
		FROM debt_payments
		WHERE debt_id = ?
		ORDER BY payment_date ASC
	`, request.ID).Scan(&payments).Error; err != nil {
		return nil, err
	}
	debt.Payments = payments

	// Ambil items kalau SALE
	if debt.ReferenceType == "SALE" {
		var items []model.DebtItemResponse
		if err := db.Raw(`
			SELECT p.name AS product_name, s.name AS size_name, sd.qty, sd.sell_price
			FROM sale_details sd
			JOIN sizes s ON sd.size_id = s.id
			JOIN products p ON s.product_sku = p.sku
			WHERE sd.sale_code = ?
		`, debt.ReferenceCode).Scan(&items).Error; err != nil {
			return nil, err
		}
		debt.Items = items
	}

	return &debt, nil
}

func (r *DebtRepository) FindBySaleCode(db *gorm.DB, debt *entity.Debt, saleCode string) error {
	return db.Where("reference_type = 'SALE' AND reference_code = ?", saleCode).Take(debt).Error
}

func (r *DebtRepository) FindBySaleCodeOrInit(db *gorm.DB, debt *entity.Debt, saleCode string) error {
	return db.Where("reference_type = 'SALE' AND reference_code = ?", saleCode).FirstOrInit(debt).Error
}

func (r *DebtRepository) FindByPurchaseCode(db *gorm.DB, debt *entity.Debt, purchaseCode string) error {
	return db.Where("reference_type = 'PURCHASE' AND reference_code = ?", purchaseCode).First(debt).Error
}

func (r *DebtRepository) FindByPurchaseCodeOrInit(db *gorm.DB, debt *entity.Debt, purchaseCode string) error {
	return db.Where("reference_type = 'PURCHASE' AND reference_code = ?", purchaseCode).FirstOrInit(debt).Error
}

func (r *DebtRepository) UpdateStatus(db *gorm.DB, id uint) error {
	return db.Model(&entity.Debt{}).
		Where("id = ?", id).
		UpdateColumn("status", "VOID").
		Error
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
