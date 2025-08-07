package repository

import (
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleSummaryRepository struct {
	Log *logrus.Logger
}

func NewSaleSummaryRepository(log *logrus.Logger) *SaleSummaryRepository {
	return &SaleSummaryRepository{
		Log: log,
	}
}

func (r *SaleSummaryRepository) BranchSalesSummary(db *gorm.DB) ([]model.BranchSalesSummaryResponse, error) {

	var results []model.BranchSalesSummaryResponse

	query := `
	SELECT 
	b.id AS branch_id,
	b.name AS branch_name,
	COALESCE(SUM(s.cash_value + s.debit_value), 0) AS total_sales,
	COALESCE(SUM(sd.qty), 0) AS total_product_sold
	FROM branches b
	LEFT JOIN sales s ON s.branch_id = b.id AND s.status = ?
	LEFT JOIN sale_details sd ON sd.sale_code = s.code AND sd.is_cancelled = false
	GROUP BY b.id, b.name
	ORDER BY b.name ASC;
	`

	if err := db.Raw(query, model.COMPLETED).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

func (r *SaleSummaryRepository) DailySalesSummaryByBranchID(db *gorm.DB, request *model.ListDailySalesSummaryRequest) ([]model.DailySalesSummaryResponse, error) {

	var summaries []model.DailySalesSummaryResponse

	query := `
	SELECT
		DATE(FROM_UNIXTIME(s.created_at / 1000)) AS date,
		SUM(s.cash_value + s.debit_value) AS total_sales,
		SUM(sd.qty) AS total_product_sold
	FROM sales s
	JOIN sale_details sd ON s.code = sd.sale_code
	WHERE s.status = ?
	  AND sd.is_cancelled = false
	  AND s.branch_id = ?
`

	params := []interface{}{model.COMPLETED, request.BranchID}

	if request.StartAt != "" {
		query += ` AND DATE(FROM_UNIXTIME(s.created_at / 1000)) >= ?`
		params = append(params, request.StartAt)
	}
	if request.EndAt != "" {
		query += ` AND DATE(FROM_UNIXTIME(s.created_at / 1000)) <= ?`
		params = append(params, request.EndAt)
	}

	query += ` GROUP BY date ORDER BY date ASC LIMIT ? OFFSET ?`
	params = append(params, request.Size, (request.Page-1)*request.Size)

	if err := db.Raw(query, params...).Scan(&summaries).Error; err != nil {
		return nil, err
	}

	return summaries, nil

}
