package model

import "time"

// type FinancialReportResponse struct {
// 	TotalPendapatan             float64 `json:"total_pendapatan"`              // dari sales (COMPLETED)
// 	TotalPengeluaranOperasional float64 `json:"total_pengeluaran_operasional"` // dari tabel expenses (optional)
// 	TotalPembelianBarang        float64 `json:"total_pembelian_barang"`        // dari purchases (COMPLETED)
// 	TotalPiutang                float64 `json:"total_piutang"`                 // sales PENDING
// 	TotalHutang                 float64 `json:"total_hutang"`                  // purchases PENDING
// 	LabaBersih                  float64 `json:"laba_bersih"`
// }

type DailyFinancialReportResponse struct {
	Date             string  `json:"date" gorm:"column:date"`
	TotalRevenue     float64 `json:"total_revenue" gorm:"column:total_revenue"`
	TotalExpenses    float64 `json:"total_expenses" gorm:"column:total_expenses"`
	NetProfit        float64 `json:"net_profit" gorm:"column:net_profit"`
	TotalDebt        float64 `json:"total_debt" gorm:"column:total_debt"`
	TotalReceivables float64 `json:"total_receivables" gorm:"column:total_receivables"`
}

type SearchDailyFinancialReportRequest struct {
	StartDate time.Time `json:"-"`
	EndDate   time.Time `json:"-"`
	BranchID  *int      `json:"-"`
}
