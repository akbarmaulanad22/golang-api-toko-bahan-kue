package repository

import (
	"fmt"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FinancialReportRepository struct {
	Log *logrus.Logger
}

func NewFinancialReportRepository(log *logrus.Logger) *FinancialReportRepository {
	return &FinancialReportRepository{
		Log: log,
	}
}

func (r *FinancialReportRepository) SearchDailyFinancialReport(
	db *gorm.DB,
	request *model.SearchDailyFinancialReportRequest,
) ([]model.DailyFinancialReportResponse, error) {
	var reports []model.DailyFinancialReportResponse

	var branchConditionSales, branchConditionExpenses, branchConditionPurchases, branchConditionDebt, branchConditionReceivables string
	params := []interface{}{}

	if request.BranchID != nil {
		branchConditionSales = "AND s.branch_id = ?"
		branchConditionExpenses = "AND e.branch_id = ?"
		branchConditionPurchases = "AND p.branch_id = ?"
		branchConditionDebt = "AND s.branch_id = ?"
		branchConditionReceivables = "AND p.branch_id = ?"
	}

	query := fmt.Sprintf(`
    SELECT
        DATE(FROM_UNIXTIME(d.tanggal / 1000)) AS date,
        COALESCE(SUM(d.total_pendapatan), 0) AS total_revenue,
        COALESCE(SUM(d.total_pengeluaran), 0) AS total_expenses,
        COALESCE(SUM(d.total_pendapatan - d.total_pengeluaran), 0) AS net_profit,
        COALESCE((
            SELECT SUM(sd.sell_price * sd.qty)
            FROM sales s
            JOIN sale_details sd ON sd.sale_code = s.code
            WHERE s.status = 'PENDING'
              AND DATE(FROM_UNIXTIME(s.paid_at / 1000)) = DATE(FROM_UNIXTIME(d.tanggal / 1000))
              %s
        ), 0) AS total_debt,
        COALESCE((
            SELECT SUM(pd.buy_price * pd.qty)
            FROM purchases p
            JOIN purchase_details pd ON pd.purchase_code = p.code
            WHERE p.status = 'PENDING'
              AND DATE(FROM_UNIXTIME(p.paid_at / 1000)) = DATE(FROM_UNIXTIME(d.tanggal / 1000))
              %s
        ), 0) AS total_receivables
    FROM (
        SELECT
            UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(s.created_at / 1000))) * 1000 AS tanggal,
            SUM(sd.sell_price * sd.qty) AS total_pendapatan,
            0 AS total_pengeluaran
        FROM sales s
        JOIN sale_details sd ON sd.sale_code = s.code
        WHERE s.status = 'COMPLETED'
          AND s.created_at BETWEEN UNIX_TIMESTAMP(?) * 1000 AND UNIX_TIMESTAMP(?) * 1000
          %s
        GROUP BY tanggal

        UNION ALL

        SELECT
            UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(e.created_at / 1000))) * 1000 AS tanggal,
            0 AS total_pendapatan,
            SUM(e.amount) AS total_pengeluaran
        FROM expenses e
        WHERE e.created_at BETWEEN UNIX_TIMESTAMP(?) * 1000 AND UNIX_TIMESTAMP(?) * 1000
          %s
        GROUP BY tanggal

        UNION ALL

        SELECT
            UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(p.created_at / 1000))) * 1000 AS tanggal,
            0 AS total_pendapatan,
            SUM(pd.buy_price * pd.qty) AS total_pengeluaran
        FROM purchases p
        JOIN purchase_details pd ON pd.purchase_code = p.code
        WHERE p.status = 'COMPLETED'
          AND p.created_at BETWEEN UNIX_TIMESTAMP(?) * 1000 AND UNIX_TIMESTAMP(?) * 1000
          %s
        GROUP BY tanggal
    ) AS d
    GROUP BY tanggal
    ORDER BY tanggal ASC;
    `, branchConditionDebt, branchConditionReceivables, branchConditionSales, branchConditionExpenses, branchConditionPurchases)

	if request.BranchID != nil {
		params = append(params,
			*request.BranchID,                                                                               // debt
			*request.BranchID,                                                                               // receivables
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), *request.BranchID, // sales
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), *request.BranchID, // expenses
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), *request.BranchID, // purchases
		)
	} else {
		params = append(params,
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), // sales
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), // expenses
			request.StartDate.Format("2006-01-02"), request.EndDate.Format("2006-01-02"), // purchases
		)
	}

	if err := db.Raw(query, params...).Scan(&reports).Error; err != nil {
		return nil, err
	}

	return reports, nil
}
