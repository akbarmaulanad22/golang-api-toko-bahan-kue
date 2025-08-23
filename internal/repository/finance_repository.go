package repository

import (
	"fmt"
	"time"
	"tokobahankue/internal/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type FinanceRepository struct {
	Log *logrus.Logger
}

func NewFinanceRepository(log *logrus.Logger) *FinanceRepository {
	return &FinanceRepository{
		Log: log,
	}
}

func (r *FinanceRepository) GetOwnerSummary(db *gorm.DB, request *model.SearchFinanceSummaryOwnerRequest) (*model.FinanceSummaryOwnerResponse, error) {
	var totalSales float64
	var totalHPP float64
	var totalExpenses int64

	// --- Total Sales ---
	if err := db.Raw(`
		SELECT COALESCE(SUM(sd.qty * sd.sell_price), 0)
		FROM sale_details sd
		JOIN sales s ON s.code = sd.sale_code
		WHERE s.status = 'COMPLETED'
		  AND sd.is_cancelled = 0
		  AND s.created_at BETWEEN ? AND ?;
	`, request.StartAt, request.EndAt).Scan(&totalSales).Error; err != nil {
		return nil, err
	}

	// --- Total HPP (proxy: total purchase cost in period) ---
	if err := db.Raw(`
		SELECT COALESCE(SUM(pd.qty * pd.buy_price), 0)
		FROM purchase_details pd
		JOIN purchases p ON p.code = pd.purchase_code
		WHERE p.status = 'COMPLETED'
		  AND pd.is_cancelled = 0
		  AND p.created_at BETWEEN ? AND ?;
	`, request.StartAt, request.EndAt).Scan(&totalHPP).Error; err != nil {
		return nil, err
	}

	// --- Total Expenses ---
	if err := db.Raw(`
		SELECT COALESCE(SUM(e.amount), 0)
		FROM expenses e
		WHERE e.created_at BETWEEN ? AND ?;
	`, request.StartAt, request.EndAt).Scan(&totalExpenses).Error; err != nil {
		return nil, err
	}

	// --- Per Branch ---
	type row struct {
		BranchID   uint
		BranchName string
		Sales      *float64
		HPP        *float64
		Expenses   *int64
	}
	var rows []row

	if err := db.Raw(`
		SELECT
		  b.id   AS branch_id,
		  b.name AS branch_name,
		  COALESCE(sales.total_sales, 0)   AS sales,
		  COALESCE(hpp.total_hpp, 0)       AS hpp,
		  COALESCE(exp.total_expenses, 0)  AS expenses
		FROM branches b
		LEFT JOIN (
		  SELECT s.branch_id, SUM(sd.qty * sd.sell_price) AS total_sales
		  FROM sales s
		  JOIN sale_details sd ON sd.sale_code = s.code
		  WHERE s.status = 'COMPLETED'
		    AND sd.is_cancelled = 0
		    AND s.created_at BETWEEN ? AND ?
		  GROUP BY s.branch_id
		) AS sales ON sales.branch_id = b.id
		LEFT JOIN (
		  SELECT p.branch_id, SUM(pd.qty * pd.buy_price) AS total_hpp
		  FROM purchases p
		  JOIN purchase_details pd ON pd.purchase_code = p.code
		  WHERE p.status = 'COMPLETED'
		    AND pd.is_cancelled = 0
		    AND p.created_at BETWEEN ? AND ?
		  GROUP BY p.branch_id
		) AS hpp ON hpp.branch_id = b.id
		LEFT JOIN (
		  SELECT e.branch_id, SUM(e.amount) AS total_expenses
		  FROM expenses e
		  WHERE e.created_at BETWEEN ? AND ?
		  GROUP BY e.branch_id
		) AS exp ON exp.branch_id = b.id
	`, request.StartAt, request.EndAt, request.StartAt, request.EndAt, request.StartAt, request.EndAt).Scan(&rows).Error; err != nil {
		return nil, err
	}

	byBranch := make([]model.BranchSummaryResponse, 0, len(rows))
	for _, r := range rows {
		sales := 0.0
		if r.Sales != nil {
			sales = *r.Sales
		}
		hpp := 0.0
		if r.HPP != nil {
			hpp = *r.HPP
		}
		exp := int64(0)
		if r.Expenses != nil {
			exp = *r.Expenses
		}

		byBranch = append(byBranch, model.BranchSummaryResponse{
			BranchID:   r.BranchID,
			BranchName: r.BranchName,
			Sales:      sales,
			HPP:        hpp,
			Expenses:   exp,
			NetProfit:  sales - hpp - float64(exp),
		})
	}

	out := &model.FinanceSummaryOwnerResponse{
		ReportType:    "owner_summary",
		Period:        fmt.Sprintf("%s to %s", formatDate(request.StartAt), formatDate(request.EndAt)),
		TotalSales:    totalSales,
		TotalHPP:      totalHPP,
		TotalExpenses: totalExpenses,
		NetProfit:     totalSales - totalHPP - float64(totalExpenses),
		ByBranch:      byBranch,
	}
	return out, nil
}

func (r *FinanceRepository) GetProfitLoss(db *gorm.DB, request *model.SearchFinanceProfitLossRequest) (*model.FinanceProfitLossResponse, error) {
	var totalSales float64
	var totalHPP float64
	var totalExpenses int64

	// --- Sales ---
	querySales := `
		SELECT COALESCE(SUM(sd.qty * sd.sell_price), 0)
		FROM sale_details sd
		JOIN sales s ON s.code = sd.sale_code
		WHERE s.status = 'COMPLETED'
		  AND sd.is_cancelled = 0
	`
	querySales, argsSales := buildFilter(querySales, request.StartAt, request.EndAt, request.Role, request.BranchID, "s")
	if err := db.Raw(querySales, argsSales...).Scan(&totalSales).Error; err != nil {
		return nil, err
	}

	// --- HPP ---
	queryHPP := `
		SELECT COALESCE(SUM(pd.qty * pd.buy_price), 0)
		FROM purchase_details pd
		JOIN purchases p ON p.code = pd.purchase_code
		WHERE p.status = 'COMPLETED'
		  AND pd.is_cancelled = 0
	`
	queryHPP, argsHPP := buildFilter(queryHPP, request.StartAt, request.EndAt, request.Role, request.BranchID, "p")
	if err := db.Raw(queryHPP, argsHPP...).Scan(&totalHPP).Error; err != nil {
		return nil, err
	}

	// --- Expenses ---
	queryExp := `
		SELECT COALESCE(SUM(e.amount), 0)
		FROM expenses e
		WHERE 1=1
	`
	queryExp, argsExp := buildFilter(queryExp, request.StartAt, request.EndAt, request.Role, request.BranchID, "e")
	if err := db.Raw(queryExp, argsExp...).Scan(&totalExpenses).Error; err != nil {
		return nil, err
	}

	// --- Response ---
	pl := &model.FinanceProfitLossResponse{
		ReportType:  "profit_loss",
		Period:      fmt.Sprintf("%s to %s", formatDate(request.StartAt), formatDate(request.EndAt)),
		Sales:       totalSales,
		HPP:         totalHPP,
		GrossProfit: totalSales - totalHPP,
		Expenses:    totalExpenses,
		NetProfit:   totalSales - totalHPP - float64(totalExpenses),
	}

	return pl, nil
}

func buildFilter(base string, startAt, endAt int64, role string, branchID uint, alias string) (string, []interface{}) {
	args := []interface{}{}
	// date filter
	if startAt > 0 && endAt > 0 {
		base += fmt.Sprintf(" AND %s.created_at BETWEEN ? AND ? ", alias)
		args = append(args, startAt, endAt)
	}
	// branch filter
	if role != "Owner" {
		base += fmt.Sprintf(" AND %s.branch_id = ? ", alias)
		args = append(args, branchID)
	}
	return base, args
}

func formatDate(ms int64) string {
	if ms == 0 {
		return ""
	}
	// ubah ms ke detik
	t := time.Unix(ms/1000, 0)
	return t.Format("2006-01-02")
}
