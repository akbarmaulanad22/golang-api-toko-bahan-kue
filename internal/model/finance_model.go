package model

type BranchSummaryResponse struct {
	BranchID   uint    `json:"branch_id"`
	BranchName string  `json:"branch_name"`
	Sales      float64 `json:"sales"`
	HPP        float64 `json:"hpp"`
	Expenses   int64   `json:"expenses"`
	NetProfit  float64 `json:"net_profit"`
}

type FinanceSummaryOwnerResponse struct {
	ReportType    string                  `json:"report_type"`
	Period        string                  `json:"period"`
	TotalSales    float64                 `json:"total_sales"`
	TotalHPP      float64                 `json:"total_hpp"`
	TotalExpenses int64                   `json:"total_expenses"`
	NetProfit     float64                 `json:"net_profit"`
	ByBranch      []BranchSummaryResponse `json:"by_branch"`
}

type GetFinanceSummaryOwnerRequest struct {
	StartAt int64 `json:"start_at"`
	EndAt   int64 `json:"end_at"`
}

type GetFinanceBasicRequest struct {
	StartAt  int64  `json:"start_at"`
	EndAt    int64  `json:"end_at"`
	BranchID uint   `json:"branch_id"`
	Role     string `json:"role"`
}

type FinanceProfitLossResponse struct {
	ReportType  string  `json:"report_type"`
	Period      string  `json:"period"`
	Sales       float64 `json:"sales"`
	HPP         float64 `json:"hpp"`
	GrossProfit float64 `json:"gross_profit"`
	Expenses    int64   `json:"expenses"`
	NetProfit   float64 `json:"net_profit"`
}

type BranchCashFlow struct {
	BranchID   uint    `json:"branch_id"`
	BranchName string  `json:"branch_name"`
	CashIn     float64 `json:"cash_in"`
	CashOut    float64 `json:"cash_out"`
	Balance    float64 `json:"balance"`
}

type FinanceCashFlowResponse struct {
	ReportType string           `json:"report_type"`
	Period     string           `json:"period"`
	CashIn     float64          `json:"cash_in"`
	CashOut    float64          `json:"cash_out"`
	Balance    float64          `json:"balance"`
	ByBranch   []BranchCashFlow `json:"by_branch,omitempty"`
}

type FinanceBalanceSheetResponse struct {
	ReportType  string      `json:"report_type"`
	AsOf        string      `json:"as_of"`
	BranchID    *uint       `json:"branch_id,omitempty"`
	BranchName  string      `json:"branch_name,omitempty"`
	Assets      Assets      `json:"assets"`
	Liabilities Liabilities `json:"liabilities"`
	Equity      Equity      `json:"equity"`
	Balance     Balance     `json:"balance"`
}

type Assets struct {
	CashAndBank        float64 `json:"cash_and_bank"`
	AccountsReceivable float64 `json:"accounts_receivable"`
	Inventory          float64 `json:"inventory"`
	TotalCurrentAssets float64 `json:"total_current_assets"`
}

type Liabilities struct {
	AccountsPayable         float64 `json:"accounts_payable"`
	TotalCurrentLiabilities float64 `json:"total_current_liabilities"`
}

type Equity struct {
	OwnerCapital     float64 `json:"owner_capital"`
	RetainedEarnings float64 `json:"retained_earnings"`
	TotalEquity      float64 `json:"total_equity"`
}

type Balance struct {
	TotalAssets           float64 `json:"total_assets"`
	LiabilitiesPlusEquity float64 `json:"liabilities_plus_equity"`
}

type GetFinanceBalanceSheetRequest struct {
	BranchID *uint  `json:"-"` // opsional (kalau owner bisa kosong artinya semua cabang)
	Role     string `json:"-"` // "Owner" atau "Admin"
	AsOf     int64  `json:"-"` // timestamp mili → akan diformat jadi "YYYY-MM-DD"
}
