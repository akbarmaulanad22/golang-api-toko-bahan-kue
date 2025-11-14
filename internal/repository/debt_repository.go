package repository

import (
	"time"
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

	// Default empty slice (bukan nil)
	debt.Items = []model.DebtItemResponse{}

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

	// Ambil items
	switch debt.ReferenceType {
	case "SALE":
		var items []model.DebtItemResponse
		if err := db.Raw(`
			SELECT p.name AS product_name,
				   s.name AS size_name,
				   sd.qty,
				   sd.sell_price
			FROM sale_details sd
			JOIN branch_inventory bi ON sd.branch_inventory_id = bi.id
			JOIN sizes s ON bi.size_id = s.id
			JOIN products p ON s.product_sku = p.sku
			WHERE sd.sale_code = ?
		`, debt.ReferenceCode).Scan(&items).Error; err != nil {
			return nil, err
		}
		debt.Items = items

	case "PURCHASE":
		var items []model.DebtItemResponse
		if err := db.Raw(`
			SELECT p.name AS product_name,
				   s.name AS size_name,
				   pd.qty,
				   pd.buy_price
			FROM purchase_details pd
			JOIN branch_inventory bi ON pd.branch_inventory_id = bi.id
			JOIN sizes s ON bi.size_id = s.id
			JOIN products p ON s.product_sku = p.sku
			WHERE pd.purchase_code = ?
		`, debt.ReferenceCode).Scan(&items).Error; err != nil {
			return nil, err
		}
		debt.Items = items
	}

	return &debt, nil
}

func (r *DebtRepository) FindBySaleCodeOrInit(db *gorm.DB, debt *entity.Debt, saleCode string) error {
	return db.Where("reference_type = 'SALE' AND reference_code = ?", saleCode).FirstOrInit(debt).Error
}

func (r *DebtRepository) FindByPurchaseCodeOrInit(db *gorm.DB, debt *entity.Debt, purchaseCode string) error {
	return db.Where("reference_type = 'PURCHASE' AND reference_code = ?", purchaseCode).FirstOrInit(debt).Error
}

func (r *DebtRepository) UpdateStatus(db *gorm.DB, id uint) error {
	return db.Model(&entity.Debt{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "VOID",
			"updated_at": time.Now().UnixMilli(),
		}).
		Error
}
