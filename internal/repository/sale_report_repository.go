package repository

import (
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SaleReportRepository struct {
	Log *logrus.Logger
}

func NewSaleReportRepository(log *logrus.Logger) *SaleReportRepository {
	return &SaleReportRepository{
		Log: log,
	}
}

// simple
func (r *SaleReportRepository) SearchDaily(db *gorm.DB, request *model.SearchSalesReportRequest) ([]model.SalesDailyReportResponse, int64, error) {
	rows := []struct {
		Date              string
		BranchID          uint
		BranchName        string
		TotalTransactions int
		TotalProductsSold int
		TotalRevenue      float64
	}{}

	query := `
		SELECT 
			DATE(FROM_UNIXTIME(s.created_at / 1000)) AS date,
			b.id AS branch_id,
			b.name AS branch_name,
			COUNT(DISTINCT s.code) AS total_transactions,
			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty ELSE 0 END) AS total_products_sold,
			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty * sd.sell_price ELSE 0 END) AS total_revenue
		FROM sales s
		JOIN branches b ON s.branch_id = b.id
		JOIN sale_details sd ON s.code = sd.sale_code
		WHERE s.status = 'COMPLETED'
	`

	args := []interface{}{}

	// filter cabang
	if request.BranchID != nil {
		query += " AND s.branch_id = ?"
		args = append(args, *request.BranchID)
	}

	// filter tanggal
	if request.StartAt > 0 && request.EndAt > 0 {
		query += " AND s.created_at BETWEEN ? AND ?"
		args = append(args, request.StartAt, request.EndAt)
	}

	query += `
		GROUP BY DATE(FROM_UNIXTIME(s.created_at / 1000)), b.id, b.name
		ORDER BY date, branch_id
	`

	if err := db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	// mapping ke response
	results := make([]model.SalesDailyReportResponse, 0, len(rows))
	for _, r := range rows {
		results = append(results, model.SalesDailyReportResponse{
			Date:              r.Date,
			BranchID:          r.BranchID,
			BranchName:        r.BranchName,
			TotalTransactions: r.TotalTransactions,
			TotalProductsSold: r.TotalProductsSold,
			TotalRevenue:      r.TotalRevenue,
		})
	}

	return results, int64(len(results)), nil
}

// untuk data sangat detail sampai utang dan perbandingan harga jual vs uang masuk
// func (r *SaleReportRepository) SearchDaily(db *gorm.DB, request *model.SearchSalesReportRequest) ([]model.SalesDailyReportResponse, int64, error) {
// 	// Step 1: Summary penjualan (transaksi, produk, revenue)
// 	summaryRows := []struct {
// 		Date              string
// 		BranchID          uint
// 		BranchName        string
// 		TotalTransactions int
// 		TotalProductsSold int
// 		TotalRevenue      float64
// 	}{}

// 	query := `
// 		SELECT
// 			DATE(FROM_UNIXTIME(s.created_at / 1000)) AS date,
// 			b.id AS branch_id,
// 			b.name AS branch_name,
// 			COUNT(DISTINCT s.code) AS total_transactions,
// 			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty ELSE 0 END) AS total_products_sold,
// 			SUM(CASE WHEN sd.is_cancelled = 0 THEN sd.qty * sd.sell_price ELSE 0 END) AS total_revenue
// 		FROM sales s
// 		JOIN branches b ON s.branch_id = b.id
// 		JOIN sale_details sd ON s.code = sd.sale_code
// 		WHERE s.status = 'COMPLETED'
// 	`

// 	args := []interface{}{}

// 	// branch filter
// 	if request.BranchID != nil {
// 		query += " AND s.branch_id = ?"
// 		args = append(args, *request.BranchID)
// 	}

// 	// date filter
// 	if !request.StartAt.IsZero() && !request.EndAt.IsZero() {
// 		query += " AND s.created_at BETWEEN ? AND ?"
// 		args = append(args, request.StartAt.UnixMilli(), request.EndAt.UnixMilli())
// 	}

// 	query += `
// 		GROUP BY DATE(FROM_UNIXTIME(s.created_at / 1000)), b.id, b.name
// 		ORDER BY date, branch_id
// 	`

// 	if err := db.Raw(query, args...).Scan(&summaryRows).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	// Step 2: Ambil breakdown payment per metode
// 	paymentRows := []struct {
// 		Date          string
// 		BranchID      uint
// 		PaymentMethod string
// 		TotalAmount   float64
// 	}{}

// 	paymentQuery := `
// 		SELECT
// 			DATE(FROM_UNIXTIME(s.created_at / 1000)) AS date,
// 			s.branch_id,
// 			sp.payment_method,
// 			SUM(sp.amount) AS total_amount
// 		FROM sales s
// 		JOIN sale_payments sp ON s.code = sp.sale_code
// 		WHERE s.status = 'COMPLETED'
// 	`

// 	paymentArgs := []interface{}{}

// 	if request.BranchID != nil {
// 		paymentQuery += " AND s.branch_id = ?"
// 		paymentArgs = append(paymentArgs, *request.BranchID)
// 	}

// 	if !request.StartAt.IsZero() && !request.EndAt.IsZero() {
// 		paymentQuery += " AND s.created_at BETWEEN ? AND ?"
// 		paymentArgs = append(paymentArgs, request.StartAt.UnixMilli(), request.EndAt.UnixMilli())
// 	}

// 	paymentQuery += `
// 		GROUP BY DATE(FROM_UNIXTIME(s.created_at / 1000)), s.branch_id, sp.payment_method
// 	`

// 	if err := db.Raw(paymentQuery, paymentArgs...).Scan(&paymentRows).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	// Step 3: Gabungkan summary + payment
// 	resultMap := make(map[string]*model.SalesDailyReportResponse)

// 	for _, row := range summaryRows {
// 		key := fmt.Sprintf("%s-%d", row.Date, row.BranchID)
// 		resultMap[key] = &model.SalesDailyReportResponse{
// 			Date:              row.Date,
// 			BranchID:          row.BranchID,
// 			BranchName:        row.BranchName,
// 			TotalTransactions: row.TotalTransactions,
// 			TotalProductsSold: row.TotalProductsSold,
// 			TotalRevenue:      row.TotalRevenue,
// 			TotalPayment:      0,
// 			TotalDebt:         0,
// 			Balance:           0,
// 			PaymentMethods:    make(map[string]float64),
// 		}
// 	}

// 	for _, row := range paymentRows {
// 		key := fmt.Sprintf("%s-%d", row.Date, row.BranchID)
// 		if summary, ok := resultMap[key]; ok {
// 			summary.PaymentMethods[row.PaymentMethod] = row.TotalAmount
// 			summary.TotalPayment += row.TotalAmount
// 		}
// 	}

// 	// Step 4: Hitung debt & balance
// 	for _, summary := range resultMap {
// 		summary.TotalDebt = summary.TotalRevenue - summary.TotalPayment
// 		summary.Balance = summary.TotalPayment - summary.TotalRevenue
// 	}

// 	// Step 5: Convert ke slice
// 	results := make([]model.SalesDailyReportResponse, 0, len(resultMap))
// 	for _, v := range resultMap {
// 		results = append(results, *v)
// 	}

// 	return results, int64(len(results)), nil
// }

func (r *SaleReportRepository) SearchTopSeller(db *gorm.DB, request *model.SearchSalesReportRequest) ([]model.SalesTopSellerReportResponse, int64, error) {
	var results []model.SalesTopSellerReportResponse

	sql := `
	SELECT 
	p.sku AS product_sku,
            p.name AS product_name,
            SUM(sd.qty) AS total_qty,
            SUM(sd.qty * sz.sell_price) AS total_omzet
			FROM sale_details sd
			JOIN sales s ON s.code = sd.sale_code
        JOIN sizes sz ON sz.id = sd.size_id
        JOIN products p ON p.sku = sz.product_sku
        WHERE s.status = 'COMPLETED'
		`

	var params []interface{}

	if request.StartAt > 0 && request.EndAt > 0 {
		sql += " AND s.created_at BETWEEN ? AND ?"
		params = append(params, request.StartAt, request.EndAt)
	}

	if request.BranchID != nil {
		sql += " AND s.branch_id = ?"
		params = append(params, *request.BranchID)
	}

	sql += " GROUP BY p.sku, p.name ORDER BY total_qty DESC"

	if err := db.Raw(sql, params...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, int64(len(results)), nil
}

func (r *SaleReportRepository) SearchCategory(db *gorm.DB, request *model.SearchSalesReportRequest) ([]model.SalesCategoryResponse, int64, error) {
	var results []model.SalesCategoryResponse

	sql := `
		SELECT 
			c.id AS category_id,
			c.name AS category_name,
			COALESCE(SUM(sd.qty), 0) AS total_qty,
			COALESCE(SUM(sd.qty * sz.sell_price), 0) AS total_omzet
		FROM sale_details sd
		JOIN sales s ON s.code = sd.sale_code
		JOIN sizes sz ON sz.id = sd.size_id
		JOIN products p ON p.sku = sz.product_sku
		JOIN categories c ON c.id = p.category_id
		WHERE s.status = 'COMPLETED'
	`

	var params []interface{}

	if request.StartAt > 0 && request.EndAt > 0 {
		sql += " AND s.created_at BETWEEN ? AND ?"
		params = append(params, request.StartAt, request.EndAt)
	}

	if request.BranchID != nil {
		sql += " AND s.branch_id = ?"
		params = append(params, *request.BranchID)
	}

	sql += " GROUP BY c.id, c.name ORDER BY total_qty DESC"

	if err := db.Raw(sql, params...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, int64(len(results)), nil
}
