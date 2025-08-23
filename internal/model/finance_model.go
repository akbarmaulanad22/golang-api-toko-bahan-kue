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

type SearchFinanceSummaryOwnerRequest struct {
	StartAt int64 `json:"start_at"`
	EndAt   int64 `json:"end_at"`
}

type SearchFinanceProfitLossRequest struct {
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
