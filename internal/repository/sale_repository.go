package repository

import (
	"tokobahankue/internal/entity"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleRepository struct {
	Repository[entity.Sale]
	Log *logrus.Logger
}

func NewSaleRepository(log *logrus.Logger) *SaleRepository {
	return &SaleRepository{
		Log: log,
	}
}

// func (r *SaleRepository) CountBySKU(db *gorm.DB, sku any) (int64, error) {
// 	var total int64
// 	err := db.Model(&entity.Sale{}).Where("sku = ?", sku).Count(&total).Error
// 	return total, err
// }

func (r *SaleRepository) FindByCode(db *gorm.DB, sale *entity.Sale, code string) error {
	return db.Preload("Branch").Where("code = ?", code).First(sale).Error
}

func (r *SaleRepository) Search(db *gorm.DB, request *model.SearchSaleRequest) ([]entity.Sale, int64, error) {
	var sales []entity.Sale
	if err := db.Scopes(r.FilterSale(request)).Offset((request.Page - 1) * request.Size).Limit(request.Size).Find(&sales).Error; err != nil {
		return nil, 0, err
	}

	var total int64 = 0
	if err := db.Model(&entity.Sale{}).Scopes(r.FilterSale(request)).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return sales, total, nil
}

func (r *SaleRepository) FilterSale(request *model.SearchSaleRequest) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if code := request.Code; code != "" {
			tx = tx.Where("code = ?", code)
		}

		if customerName := request.CustomerName; customerName != "" {
			customerName = "%" + customerName + "%"
			tx = tx.Where("customer_name LIKE ?", customerName)
		}

		if status := request.Status; status != "" {
			tx = tx.Where("status = ?", status)
		}

		startAt := request.StartAt
		endAt := request.EndAt
		if startAt != 0 && endAt != 0 {
			tx = tx.Where("paid_at BETWEEN ? AND ? OR created_at BETWEEN ? AND ?", startAt, endAt, startAt, endAt)
		}

		return tx
	}
}

func (r *SaleRepository) SearchReports(db *gorm.DB, request *model.SearchSaleReportRequest) ([]model.SaleReportResponse, int64, error) {

	var reports []model.SaleReportResponse

	query := `
    SELECT
        FROM_UNIXTIME(s.created_at / 1000) AS created_at,
        b.id,
        b.name AS branch_name,
        s.code AS sale_code,
        s.customer_name,
        p.name AS product_name,
        sd.qty,
        sz.sell_price,
        (sd.qty * sz.sell_price) AS total_price
		FROM sales s
		JOIN branches b ON s.branch_id = b.id
		JOIN sale_details sd ON s.code = sd.sale_code AND sd.is_cancelled = false
		JOIN sizes sz ON sd.size_id = sz.id
		JOIN products p ON sz.product_sku = p.sku
    WHERE s.status = ?
`

	params := []interface{}{model.COMPLETED}

	// Filter tanggal
	if request.StartAt != "" && request.EndAt != "" {
		query += " AND DATE(FROM_UNIXTIME(s.created_at / 1000)) BETWEEN ? AND ?"
		params = append(params, request.StartAt, request.EndAt)
	}

	// Filter customer name
	if request.Search != "" {
		search := "%" + request.Search + "%"
		query += "AND (s.customer_name LIKE ? OR p.name LIKE ? OR s.code LIKE ?)"
		params = append(params, search, search, search)
	}

	// Filter branch
	if request.BranchID != 0 {
		query += " AND b.id = ?"
		params = append(params, request.BranchID)
	}

	// Urutkan berdasarkan tanggal
	query += " ORDER BY created_at ASC"

	if err := db.Raw(query, params...).Scan(&reports).Error; err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS sub"
	var total int64 = 0
	if err := db.Raw(countQuery, params...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil

}

func (r *SaleRepository) SummaryAllBranch(db *gorm.DB) ([]model.BranchSalesReportResponse, error) {

	var branchSalesReport []model.BranchSalesReportResponse

	query := `
		SELECT 
            b.name AS branch_name,
            SUM(sd.qty * sz.sell_price) AS total_sales
        FROM sales s
        JOIN branches b ON s.branch_id = b.id
        JOIN sale_details sd ON s.code = sd.sale_code AND sd.is_cancelled = false
        JOIN sizes sz ON sd.size_id = sz.id
        GROUP BY b.name

        UNION ALL

        SELECT 
            'Total Semua Cabang' AS branch_name,
            SUM(sd.qty * sz.sell_price) AS total_sales
        FROM sales s
        JOIN branches b ON s.branch_id = b.id
        JOIN sale_details sd ON s.code = sd.sale_code AND sd.is_cancelled = false
        JOIN sizes sz ON sd.size_id = sz.id
        ORDER BY branch_name;
	`

	if err := db.Raw(query).Scan(&branchSalesReport).Error; err != nil {
		return nil, err
	}

	return branchSalesReport, nil
}

func (r *SaleRepository) FindhBestSellingProductsGlobal(db *gorm.DB) ([]model.BestSellingProductResponse, error) {

	query := `
        SELECT 
			p.sku,
			p.name AS product_name,
			SUM(sd.qty) AS total_qty,
			SUM(sd.qty * sz.sell_price) AS total_sales
		FROM sale_details sd
		JOIN sales s ON sd.sale_code = s.code
		JOIN sizes sz ON sd.size_id = sz.id
		JOIN products p ON sz.product_sku = p.sku
		WHERE s.status = 'COMPLETED' AND sd.is_cancelled = 0
		GROUP BY p.sku, p.name
		ORDER BY total_qty DESC
		LIMIT 10;
    `
	var result []model.BestSellingProductResponse
	err := db.Raw(query).Scan(&result).Error
	return result, err
}

func (r *SaleRepository) FindhBestSellingProductsByBranchID(db *gorm.DB, request *model.ListBestSellingProductRequest) ([]model.BestSellingProductResponse, error) {

	query := `
        SELECT 
			p.sku,
			p.name AS product_name,
			SUM(sd.qty) AS total_qty,
			SUM(sd.qty * sz.sell_price) AS total_sales
		FROM sale_details sd
		JOIN sales s ON sd.sale_code = s.code
		JOIN sizes sz ON sd.size_id = sz.id
		JOIN products p ON sz.product_sku = p.sku
		WHERE s.status = 'COMPLETED' 
		AND sd.is_cancelled = 0
		AND s.branch_id = ?
		GROUP BY p.sku, p.name
		ORDER BY total_qty DESC
		LIMIT 10;
    `
	var result []model.BestSellingProductResponse
	err := db.Raw(query, request.BranchID).Scan(&result).Error
	return result, err
}
