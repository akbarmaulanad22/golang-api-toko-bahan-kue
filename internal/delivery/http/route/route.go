package route

import (
	"tokobahankue/internal/delivery/http"
	"tokobahankue/internal/delivery/http/middleware"

	"github.com/gorilla/mux"
)

type RouteConfig struct {
	// router
	Router *mux.Router

	// middleware
	AuthMiddleware mux.MiddlewareFunc
	RoleMiddleware middleware.MiddlewareCustom

	// all field controller

	// master data
	BranchController      *http.BranchController
	RoleController        *http.RoleController
	CategoryController    *http.CategoryController
	DistributorController *http.DistributorController
	// end master data

	DashboardController           *http.DashboardController
	ProductController             *http.ProductController
	UserController                *http.UserController
	SizeController                *http.SizeController
	SaleController                *http.SaleController
	PurchaseController            *http.PurchaseController
	SaleReportController          *http.SaleReportController
	ExpenseController             *http.ExpenseController
	CapitalController             *http.CapitalController
	CashBankTransactionController *http.CashBankTransactionController
	DebtController                *http.DebtController
	DebtPaymentController         *http.DebtPaymentController
	FinanceController             *http.FinanceController
	BranchInventoryController     *http.BranchInventoryController
	InventoryMovementController   *http.InventoryMovementController
	SaleDetailController          *http.SaleDetailController
	PurchaseDetailController      *http.PurchaseDetailController
	PurchaseReportController      *http.PurchaseReportController
	StockOpnameController         *http.StockOpnameController
}

func (route *RouteConfig) Setup() {
	route.SetupGuestRoute()
	route.SetupAuthRoute()
}

func (route *RouteConfig) SetupGuestRoute() {
	// routes that do not require authentication

	// auth
	route.Router.HandleFunc("/gate/auth/login", middleware.WithErrorHandler(route.UserController.Login)).Methods("POST")
	route.Router.HandleFunc("/gate/auth/register", middleware.WithErrorHandler(route.UserController.Register)).Methods("POST")
	// end auth
}

func (route *RouteConfig) SetupAuthRoute() {

	// base path
	apiRouter := route.Router.PathPrefix("/api/v1/").Subrouter()

	// Subrouter dengan AuthMiddleware
	authRouter := apiRouter.NewRoute().Subrouter()
	authRouter.Use(route.AuthMiddleware)

	// Subrouter khusus role tertentu (harus turun dari authRouter)
	adminRouter := authRouter.NewRoute().Subrouter()
	adminRouter.Use(route.RoleMiddleware("admin"))

	ownerRouter := authRouter.NewRoute().Subrouter()
	ownerRouter.Use(route.RoleMiddleware("owner"))

	ownerOrAdminRouter := authRouter.NewRoute().Subrouter()
	ownerOrAdminRouter.Use(route.RoleMiddleware("owner", "admin"))

	cashierOrAdminRouter := authRouter.NewRoute().Subrouter()
	cashierOrAdminRouter.Use(route.RoleMiddleware("cashier", "admin"))

	gateRouter := route.Router.NewRoute().Subrouter()
	gateRouter.Use(route.AuthMiddleware)
	gateRouter.HandleFunc("/gate/auth/logout", middleware.WithErrorHandler(route.UserController.Logout)).Methods("POST")
	gateRouter.HandleFunc("/gate/auth/me", middleware.WithErrorHandler(route.UserController.Current)).Methods("GET")

	// master data

	// cabang
	ownerRouter.HandleFunc("/branches", middleware.WithErrorHandler(route.BranchController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/branches", middleware.WithErrorHandler(route.BranchController.List)).Methods("GET")
	ownerRouter.HandleFunc("/branches/{id}", middleware.WithErrorHandler(route.BranchController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/branches/{id}", middleware.WithErrorHandler(route.BranchController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/branches/{id}", middleware.WithErrorHandler(route.BranchController.Get)).Methods("GET")

	// jabatan/posisi
	ownerRouter.HandleFunc("/roles", middleware.WithErrorHandler(route.RoleController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/roles", middleware.WithErrorHandler(route.RoleController.List)).Methods("GET")
	ownerRouter.HandleFunc("/roles/{id}", middleware.WithErrorHandler(route.RoleController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/roles/{id}", middleware.WithErrorHandler(route.RoleController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/roles/{id}", middleware.WithErrorHandler(route.RoleController.Get)).Methods("GET")

	// kategori
	ownerRouter.HandleFunc("/categories", middleware.WithErrorHandler(route.CategoryController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/categories", middleware.WithErrorHandler(route.CategoryController.List)).Methods("GET")
	ownerRouter.HandleFunc("/categories/{id}", middleware.WithErrorHandler(route.CategoryController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/categories/{id}", middleware.WithErrorHandler(route.CategoryController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/categories/{id}", middleware.WithErrorHandler(route.CategoryController.Get)).Methods("GET")

	// distributor
	authRouter.HandleFunc("/distributors", middleware.WithErrorHandler(route.DistributorController.List)).Methods("GET")
	ownerRouter.HandleFunc("/distributors", middleware.WithErrorHandler(route.DistributorController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/distributors/{id}", middleware.WithErrorHandler(route.DistributorController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/distributors/{id}", middleware.WithErrorHandler(route.DistributorController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/distributors/{id}", middleware.WithErrorHandler(route.DistributorController.Get)).Methods("GET")
	// end master data

	// dashboard
	ownerRouter.HandleFunc("/dashboard", middleware.WithErrorHandler(route.DashboardController.Get)).Methods("GET")

	// produk
	ownerRouter.HandleFunc("/products", middleware.WithErrorHandler(route.ProductController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/products", middleware.WithErrorHandler(route.ProductController.List)).Methods("GET")
	ownerRouter.HandleFunc("/products/{sku}", middleware.WithErrorHandler(route.ProductController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/products/{sku}", middleware.WithErrorHandler(route.ProductController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/products/{sku}", middleware.WithErrorHandler(route.ProductController.Get)).Methods("GET")

	// karyawan/pengguna aplikasi
	ownerOrAdminRouter.HandleFunc("/users", middleware.WithErrorHandler(route.UserController.List)).Methods("GET")
	ownerRouter.HandleFunc("/users", middleware.WithErrorHandler(route.UserController.Register)).Methods("POST")
	ownerRouter.HandleFunc("/users/{username}", middleware.WithErrorHandler(route.UserController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/users/{username}", middleware.WithErrorHandler(route.UserController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/users/{username}", middleware.WithErrorHandler(route.UserController.Get)).Methods("GET")

	// ukuran produk
	ownerRouter.HandleFunc("/products/{productSKU}/sizes", middleware.WithErrorHandler(route.SizeController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/products/{productSKU}/sizes", middleware.WithErrorHandler(route.SizeController.List)).Methods("GET")
	ownerRouter.HandleFunc("/products/{productSKU}/sizes/{id}", middleware.WithErrorHandler(route.SizeController.Update)).Methods("PUT")
	ownerRouter.HandleFunc("/products/{productSKU}/sizes/{id}", middleware.WithErrorHandler(route.SizeController.Delete)).Methods("DELETE")
	ownerRouter.HandleFunc("/products/{productSKU}/sizes/{id}", middleware.WithErrorHandler(route.SizeController.Get)).Methods("GET")

	// POS (order barang keluar)
	authRouter.HandleFunc("/sales", middleware.WithErrorHandler(route.SaleController.List)).Methods("GET")
	authRouter.HandleFunc("/sales/{code}", middleware.WithErrorHandler(route.SaleController.Get)).Methods("GET")
	cashierOrAdminRouter.HandleFunc("/sales", middleware.WithErrorHandler(route.SaleController.Create)).Methods("POST")
	cashierOrAdminRouter.HandleFunc("/sales/{code}", middleware.WithErrorHandler(route.SaleController.Cancel)).Methods("DELETE")
	cashierOrAdminRouter.HandleFunc("/sales/{code}/cancel/{id}", middleware.WithErrorHandler(route.SaleDetailController.Cancel)).Methods("DELETE")

	// POS (order barang masuk)
	authRouter.HandleFunc("/purchases", middleware.WithErrorHandler(route.PurchaseController.List)).Methods("GET")
	authRouter.HandleFunc("/purchases/{code}", middleware.WithErrorHandler(route.PurchaseController.Get)).Methods("GET")
	cashierOrAdminRouter.HandleFunc("/purchases", middleware.WithErrorHandler(route.PurchaseController.Create)).Methods("POST")
	cashierOrAdminRouter.HandleFunc("/purchases/{code}", middleware.WithErrorHandler(route.PurchaseController.Cancel)).Methods("DELETE")
	cashierOrAdminRouter.HandleFunc("/purchases/{code}/cancel/{id}", middleware.WithErrorHandler(route.PurchaseDetailController.Cancel)).Methods("DELETE")

	// laporan barang keluar [ list per tanggal ]
	ownerOrAdminRouter.HandleFunc("/sales-reports/daily", middleware.WithErrorHandler(route.SaleReportController.ListDaily)).Methods("GET")

	// laporan keseluruhan barang keluar [ list barang terlaris ]
	ownerOrAdminRouter.HandleFunc("/sales-reports/top-seller-products", middleware.WithErrorHandler(route.SaleReportController.ListTopSeller)).Methods("GET")

	// laporan keseluruhan barang keluar [ list barang terlaris per category ]
	ownerOrAdminRouter.HandleFunc("/sales-reports/top-seller-categories", middleware.WithErrorHandler(route.SaleReportController.ListCategory)).Methods("GET")

	// laporan barang masuk [ list per tanggal ]
	ownerOrAdminRouter.HandleFunc("/purchases-reports/daily", middleware.WithErrorHandler(route.PurchaseReportController.ListDaily)).Methods("GET")

	// pengeluaran
	ownerOrAdminRouter.HandleFunc("/expenses/consolidated", middleware.WithErrorHandler(route.ExpenseController.ConsolidatedReport)).Methods("GET")
	ownerOrAdminRouter.HandleFunc("/expenses", middleware.WithErrorHandler(route.ExpenseController.Create)).Methods("POST")
	ownerOrAdminRouter.HandleFunc("/expenses", middleware.WithErrorHandler(route.ExpenseController.List)).Methods("GET")
	ownerOrAdminRouter.HandleFunc("/expenses/{id}", middleware.WithErrorHandler(route.ExpenseController.Update)).Methods("PUT")
	ownerOrAdminRouter.HandleFunc("/expenses/{id}", middleware.WithErrorHandler(route.ExpenseController.Delete)).Methods("DELETE")

	// pencatatan modal masuk/keluar
	// authRouter.HandleFunc("/capitals/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	ownerOrAdminRouter.HandleFunc("/capitals", middleware.WithErrorHandler(route.CapitalController.Create)).Methods("POST")
	ownerOrAdminRouter.HandleFunc("/capitals", middleware.WithErrorHandler(route.CapitalController.List)).Methods("GET")
	ownerOrAdminRouter.HandleFunc("/capitals/{id}", middleware.WithErrorHandler(route.CapitalController.Update)).Methods("PUT")
	ownerOrAdminRouter.HandleFunc("/capitals/{id}", middleware.WithErrorHandler(route.CapitalController.Delete)).Methods("DELETE")

	// pencatatan penerimaan/pengeluaran uang
	// ownerOrAdminRouter.HandleFunc("/cash-bank-transactions/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	ownerOrAdminRouter.HandleFunc("/cash-bank-transactions", middleware.WithErrorHandler(route.CashBankTransactionController.List)).Methods("GET")

	// utang / piutang
	// ownerOrAdminRouter.HandleFunc("/debt/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	authRouter.HandleFunc("/debt", middleware.WithErrorHandler(route.DebtController.List)).Methods("GET")
	authRouter.HandleFunc("/debt/{id}", middleware.WithErrorHandler(route.DebtController.Get)).Methods("GET")
	cashierOrAdminRouter.HandleFunc("/debt/{debtID}/payments", middleware.WithErrorHandler(route.DebtPaymentController.Create)).Methods("POST")
	cashierOrAdminRouter.HandleFunc("/debt/{debtID}/payments/{id}", middleware.WithErrorHandler(route.DebtPaymentController.Delete)).Methods("DELETE")

	// ringkasan laporan keuangan [ owner only ]
	ownerRouter.HandleFunc("/finance-report/summary", middleware.WithErrorHandler(route.FinanceController.GetSummary)).Methods("GET")
	// laporan keuangan laba rugi [ owner, admin cabang ]
	ownerOrAdminRouter.HandleFunc("/finance-report/profit-loss", middleware.WithErrorHandler(route.FinanceController.GetProfitLoss)).Methods("GET")
	// laporan keuangan arus kas [ owner, admin cabang ]
	ownerOrAdminRouter.HandleFunc("/finance-report/cashflow", middleware.WithErrorHandler(route.FinanceController.GetCashFlow)).Methods("GET")
	// laporan keuangan neraca [ owner, admin cabang ]
	ownerOrAdminRouter.HandleFunc("/finance-report/balance-sheet", middleware.WithErrorHandler(route.FinanceController.GetBalanceSheet)).Methods("GET")

	// stok barang
	authRouter.HandleFunc("/branch-inventory", middleware.WithErrorHandler(route.BranchInventoryController.List)).Methods("GET")
	ownerRouter.HandleFunc("/branch-inventory", middleware.WithErrorHandler(route.BranchInventoryController.Create)).Methods("POST")
	ownerRouter.HandleFunc("/branch-inventory/{id}", middleware.WithErrorHandler(route.BranchInventoryController.Delete)).Methods("DELETE")

	// pergerakan stok barang masuk/keluar
	// ownerOrAdminRouter.HandleFunc("/inventory-movement/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	authRouter.HandleFunc("/inventory-movement", middleware.WithErrorHandler(route.InventoryMovementController.List)).Methods("GET")
	authRouter.HandleFunc("/inventory-movement", middleware.WithErrorHandler(route.InventoryMovementController.Create)).Methods("POST")
	// [owner only]
	ownerRouter.HandleFunc("/inventory-movement/summary", middleware.WithErrorHandler(route.InventoryMovementController.Summary)).Methods("GET")

	// cek stok fisik dan digital
	authRouter.HandleFunc("/stock-opname", middleware.WithErrorHandler(route.StockOpnameController.List)).Methods("GET")
	authRouter.HandleFunc("/stock-opname", middleware.WithErrorHandler(route.StockOpnameController.Create)).Methods("POST")
	authRouter.HandleFunc("/stock-opname/{id}", middleware.WithErrorHandler(route.StockOpnameController.Update)).Methods("PUT")
	authRouter.HandleFunc("/stock-opname/{id}", middleware.WithErrorHandler(route.StockOpnameController.Delete)).Methods("DELETE")
	authRouter.HandleFunc("/stock-opname/{id}", middleware.WithErrorHandler(route.StockOpnameController.Get)).Methods("GET")

}
