package repository

import (
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PurchaseReportRepository struct {
	Log *logrus.Logger
}

func NewPurchaseReportRepository(log *logrus.Logger) *PurchaseReportRepository {
	return &PurchaseReportRepository{
		Log: log,
	}
}

func (r *PurchaseReportRepository) SearchDaily(db *gorm.DB, request *model.SearchPurchasesReportRequest) ([]model.PurchasesDailyReportResponse, int64, error) {
	results := []model.PurchasesDailyReportResponse{}
	args := []interface{}{}

	baseQuery := `
		FROM purchases s
		JOIN branches b ON s.branch_id = b.id
		JOIN purchase_details sd ON s.code = sd.purchase_code
		WHERE s.status = 'COMPLETED'
	`

	// filter cabang
	if request.BranchID != nil {
		baseQuery += " AND s.branch_id = ?"
		args = append(args, *request.BranchID)
	}

	// filter tanggal
	if request.StartAt > 0 && request.EndAt > 0 {
		baseQuery += " AND s.created_at BETWEEN ? AND ?"
		args = append(args, request.StartAt, request.EndAt)
	}

	// hitung total rows (sebelum limit)
	var total int64
	countQuery := "SELECT COUNT(DISTINCT DATE(FROM_UNIXTIME(s.created_at / 1000)), b.id) " + baseQuery
	if err := db.Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// query utama dengan pagination
	query := `
		SELECT 
			DATE(FROM_UNIXTIME(s.created_at / 1000)) AS date,
			b.id AS branch_id,
			b.name AS branch_name,
			COUNT(DISTINCT s.code) AS total_transactions,
			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty ELSE 0 END) AS total_products_buy,
			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty * sd.buy_price ELSE 0 END) AS total_purchases
	` + baseQuery + `
		GROUP BY DATE(FROM_UNIXTIME(s.created_at / 1000)), b.id, b.name
		ORDER BY date, branch_id
		LIMIT ? OFFSET ?
	`

	limit := request.Size
	offset := (request.Page - 1) * request.Size
	args = append(args, limit, offset)

	if err := db.Raw(query, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}
