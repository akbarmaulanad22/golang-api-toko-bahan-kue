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

func (r *FinanceRepository) GetOwnerSummary(db *gorm.DB, request *model.GetFinanceSummaryOwnerRequest) (*model.FinanceSummaryOwnerResponse, error) {
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

func (r *FinanceRepository) GetProfitLoss(db *gorm.DB, request *model.GetFinanceBasicRequest) (*model.FinanceProfitLossResponse, error) {
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

func (r *FinanceRepository) GetCashFlow(db *gorm.DB, request *model.GetFinanceBasicRequest) (*model.FinanceCashFlowResponse, error) {
	var cashIn, cashOut, balance float64

	// --- Query Total ---
	query := `
		SELECT 
		  COALESCE(SUM(CASE WHEN cbt.type = 'IN' THEN cbt.amount ELSE 0 END), 0) AS cash_in,
		  COALESCE(SUM(CASE WHEN cbt.type = 'OUT' THEN cbt.amount ELSE 0 END), 0) AS cash_out,
		  COALESCE(SUM(CASE WHEN cbt.type = 'IN' THEN cbt.amount ELSE -cbt.amount END), 0) AS balance
		FROM cash_bank_transactions cbt
		WHERE 1=1
	`

	args := []interface{}{}
	if request.StartAt > 0 && request.EndAt > 0 {
		query += " AND cbt.created_at BETWEEN ? AND ? "
		args = append(args, request.StartAt, request.EndAt)
	}
	if request.Role != "Owner" {
		query += " AND cbt.branch_id = ? "
		args = append(args, request.BranchID)
	}

	if err := db.Raw(query, args...).Row().Scan(&cashIn, &cashOut, &balance); err != nil {
		return nil, err
	}

	resp := &model.FinanceCashFlowResponse{
		ReportType: "cash_flow",
		Period:     fmt.Sprintf("%s to %s", formatDate(request.StartAt), formatDate(request.EndAt)),
		CashIn:     cashIn,
		CashOut:    cashOut,
		Balance:    balance,
	}

	// --- Owner: breakdown per cabang ---
	if request.Role == "Owner" {
		var byBranch []model.BranchCashFlow
		queryBranch := `
			SELECT 
			  cbt.branch_id,
			  b.name AS branch_name,
			  COALESCE(SUM(CASE WHEN cbt.type = 'IN' THEN cbt.amount ELSE 0 END), 0) AS cash_in,
			  COALESCE(SUM(CASE WHEN cbt.type = 'OUT' THEN cbt.amount ELSE 0 END), 0) AS cash_out,
			  COALESCE(SUM(CASE WHEN cbt.type = 'IN' THEN cbt.amount ELSE -cbt.amount END), 0) AS balance
			FROM cash_bank_transactions cbt
			JOIN branches b ON b.id = cbt.branch_id
			WHERE 1=1
		`

		argsBranch := []interface{}{}
		if request.StartAt > 0 && request.EndAt > 0 {
			queryBranch += " AND cbt.created_at BETWEEN ? AND ? "
			argsBranch = append(argsBranch, request.StartAt, request.EndAt)
		}

		queryBranch += " GROUP BY cbt.branch_id, b.name ORDER BY b.name"

		if err := db.Raw(queryBranch, argsBranch...).Scan(&byBranch).Error; err != nil {
			return nil, err
		}

		resp.ByBranch = byBranch
	}

	return resp, nil
}

func (r *FinanceRepository) GetBalanceSheet(db *gorm.DB, request *model.GetFinanceBalanceSheetRequest) (*model.FinanceBalanceSheetResponse, error) {
	// Konversi timestamp ke string date (asOf date)
	asOfDate := time.UnixMilli(request.AsOf).Format("2006-01-02")

	// --- Cash & Bank ---
	var cashBank float64
	queryCash := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM cash_bank_transactions
		WHERE created_at <= ?`
	argsCash := []interface{}{request.AsOf}
	if request.Role != "Owner" || request.BranchID != nil {
		queryCash += " AND branch_id = ?"
		argsCash = append(argsCash, request.BranchID)
	}
	if err := db.Raw(queryCash, argsCash...).Scan(&cashBank).Error; err != nil {
		return nil, err
	}

	// --- Accounts Receivable (piutang dari SALE) ---
	var receivable float64
	if request.Role == "Owner" && request.BranchID == nil {
		// Semua cabang
		if err := db.Raw(`
			SELECT COALESCE(SUM(d.total_amount - d.paid_amount), 0)
			FROM debts d
			WHERE d.reference_type = 'SALE'
			  AND d.status = 'UNPAID'
			  AND d.due_date <= ?`,
			request.AsOf).Scan(&receivable).Error; err != nil {
			return nil, err
		}
	} else {
		// Filter cabang
		if err := db.Raw(`
			SELECT COALESCE(SUM(d.total_amount - d.paid_amount), 0)
			FROM debts d
			JOIN sales s ON d.reference_code = s.code
			WHERE d.reference_type = 'SALE'
			  AND d.status = 'UNPAID'
			  AND d.due_date <= ?
			  AND s.branch_id = ?`,
			request.AsOf, request.BranchID).Scan(&receivable).Error; err != nil {
			return nil, err
		}
	}

	// --- Accounts Payable (utang dari PURCHASE) ---
	var payable float64
	if request.Role == "Owner" && request.BranchID == nil {
		if err := db.Raw(`
			SELECT COALESCE(SUM(d.total_amount - d.paid_amount), 0)
			FROM debts d
			WHERE d.reference_type = 'PURCHASE'
			  AND d.status = 'UNPAID'
			  AND d.due_date <= ?`,
			request.AsOf).Scan(&payable).Error; err != nil {
			return nil, err
		}
	} else {
		if err := db.Raw(`
			SELECT COALESCE(SUM(d.total_amount - d.paid_amount), 0)
			FROM debts d
			JOIN purchases p ON d.reference_code = p.code
			WHERE d.reference_type = 'PURCHASE'
			  AND d.status = 'UNPAID'
			  AND d.due_date <= ?
			  AND p.branch_id = ?`,
			request.AsOf, request.BranchID).Scan(&payable).Error; err != nil {
			return nil, err
		}
	}

	// --- Inventory ---
	var inventory float64
	queryInv := `
		SELECT COALESCE(SUM(bi.stock * s.buy_price), 0)
		FROM branch_inventory bi
		JOIN sizes s ON bi.size_id = s.id
		WHERE 1=1`
	argsInv := []interface{}{}
	if request.Role != "Owner" || request.BranchID != nil {
		queryInv += " AND bi.branch_id = ?"
		argsInv = append(argsInv, request.BranchID)
	}
	if err := db.Raw(queryInv, argsInv...).Scan(&inventory).Error; err != nil {
		return nil, err
	}

	// --- Owner Capital ---
	var capital float64
	queryCap := `
		SELECT COALESCE(SUM(amount), 0)
		FROM capitals
		WHERE created_at <= ?`
	argsCap := []interface{}{request.AsOf}
	if request.Role != "Owner" || request.BranchID != nil {
		queryCap += " AND branch_id = ?"
		argsCap = append(argsCap, request.BranchID)
	}
	if err := db.Raw(queryCap, argsCap...).Scan(&capital).Error; err != nil {
		return nil, err
	}

	// --- Retained Earnings (Laba ditahan) ---
	var retainedEarnings float64

	salesCond := ""
	purchCond := ""
	expCond := ""
	argsRE := []interface{}{request.AsOf, request.AsOf, request.AsOf}

	if request.Role != "Owner" || request.BranchID != nil {
		salesCond = " AND s.branch_id = ?"
		purchCond = " AND p.branch_id = ?"
		expCond = " AND e.branch_id = ?"
		argsRE = append(argsRE, request.BranchID, request.BranchID, request.BranchID)
	}

	queryRE := fmt.Sprintf(`
	SELECT 
		COALESCE((
			(SELECT SUM(sd.qty * sd.sell_price) 
			 FROM sale_details sd 
			 JOIN sales s ON s.code = sd.sale_code
			 WHERE s.status = 'COMPLETED' 
			   AND s.created_at <= ? %s)
			-
			(SELECT SUM(pd.qty * pd.buy_price)
			 FROM purchase_details pd
			 JOIN purchases p ON p.code = pd.purchase_code
			 WHERE p.status = 'COMPLETED' 
			   AND p.created_at <= ? %s)
			-
			(SELECT SUM(amount)
			 FROM expenses e
			 WHERE e.created_at <= ? %s)
		),0)`, salesCond, purchCond, expCond)

	if err := db.Raw(queryRE, argsRE...).Scan(&retainedEarnings).Error; err != nil {
		return nil, err
	}

	// --- Build response ---
	resp := &model.FinanceBalanceSheetResponse{
		ReportType: "balance_sheet",
		AsOf:       asOfDate,
		BranchID:   nil,
		BranchName: "All Branches",
		Assets: model.Assets{
			CashAndBank:        cashBank,
			AccountsReceivable: receivable,
			Inventory:          inventory,
			TotalCurrentAssets: cashBank + receivable + inventory,
		},
		Liabilities: model.Liabilities{
			AccountsPayable:         payable,
			TotalCurrentLiabilities: payable,
		},
		Equity: model.Equity{
			OwnerCapital:     capital,
			RetainedEarnings: retainedEarnings,
			TotalEquity:      capital + retainedEarnings,
		},
	}
	resp.Balance = model.Balance{
		TotalAssets:           resp.Assets.TotalCurrentAssets,
		LiabilitiesPlusEquity: resp.Liabilities.TotalCurrentLiabilities + resp.Equity.TotalEquity,
	}

	// Kalau filter cabang, ganti branch_id dan branch_name
	if request.Role != "Owner" || request.BranchID != nil {
		resp.BranchID = request.BranchID
		var branchName string
		if err := db.Raw("SELECT name FROM branches WHERE id = ?", request.BranchID).Scan(&branchName).Error; err == nil {
			resp.BranchName = branchName
		}
	}

	return resp, nil
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
