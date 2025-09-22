package route

import (
	"tokobahankue/internal/delivery/http"

	"github.com/gorilla/mux"
)

type RouteConfig struct {
	// router
	Router *mux.Router

	// middleware
	AuthMiddleware mux.MiddlewareFunc

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
}

func (route *RouteConfig) Setup() {
	route.SetupGuestRoute()
	route.SetupAuthRoute()
}

func (route *RouteConfig) SetupGuestRoute() {
	// routes that do not require authentication

	// auth
	route.Router.HandleFunc("/gate/auth/login", route.UserController.Login).Methods("POST")
	route.Router.HandleFunc("/gate/auth/register", route.UserController.Register).Methods("POST")
	// end auth
}

func (route *RouteConfig) SetupAuthRoute() {

	// Buat subrouter khusus untuk route yang butuh auth
	authRouter := route.Router.PathPrefix("/").Subrouter()
	authRouter.Use(route.AuthMiddleware)

	// logout
	authRouter.HandleFunc("/gate/auth/logout", route.UserController.Logout).Methods("POST")

	// profile
	authRouter.HandleFunc("/gate/auth/me", route.UserController.Current).Methods("GET")

	// base path
	authRouter = route.Router.PathPrefix("/api/v1/").Subrouter()
	authRouter.Use(route.AuthMiddleware)

	// master data

	// cabang
	authRouter.HandleFunc("/branches", route.BranchController.Create).Methods("POST")
	authRouter.HandleFunc("/branches", route.BranchController.List).Methods("GET")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Update).Methods("PUT")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Get).Methods("GET")

	// jabatan/posisi
	authRouter.HandleFunc("/roles", route.RoleController.Create).Methods("POST")
	authRouter.HandleFunc("/roles", route.RoleController.List).Methods("GET")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Update).Methods("PUT")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Get).Methods("GET")

	// kategori
	authRouter.HandleFunc("/categories", route.CategoryController.Create).Methods("POST")
	authRouter.HandleFunc("/categories", route.CategoryController.List).Methods("GET")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Update).Methods("PUT")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Get).Methods("GET")

	// distributor
	authRouter.HandleFunc("/distributors", route.DistributorController.Create).Methods("POST")
	authRouter.HandleFunc("/distributors", route.DistributorController.List).Methods("GET")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Update).Methods("PUT")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Get).Methods("GET")
	// end master data

	// dashboard
	authRouter.HandleFunc("/dashboard", route.DashboardController.Get).Methods("GET")

	// produk
	authRouter.HandleFunc("/products", route.ProductController.Create).Methods("POST")
	authRouter.HandleFunc("/products", route.ProductController.List).Methods("GET")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Update).Methods("PUT")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Get).Methods("GET")

	// karyawan/pengguna aplikasi
	authRouter.HandleFunc("/users", route.UserController.Register).Methods("POST")
	authRouter.HandleFunc("/users", route.UserController.List).Methods("GET")
	authRouter.HandleFunc("/users/{username}", route.UserController.Update).Methods("PUT")
	authRouter.HandleFunc("/users/{username}", route.UserController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/users/{username}", route.UserController.Get).Methods("GET")

	// ukuran produk
	authRouter.HandleFunc("/products/{productSKU}/sizes", route.SizeController.Create).Methods("POST")
	authRouter.HandleFunc("/products/{productSKU}/sizes", route.SizeController.List).Methods("GET")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Update).Methods("PUT")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Get).Methods("GET")

	// POS (order barang keluar)
	authRouter.HandleFunc("/sales", route.SaleController.Create).Methods("POST")
	authRouter.HandleFunc("/sales", route.SaleController.List).Methods("GET")
	authRouter.HandleFunc("/sales/{code}", route.SaleController.Cancel).Methods("DELETE")
	authRouter.HandleFunc("/sales/{code}", route.SaleController.Get).Methods("GET")
	authRouter.HandleFunc("/sales/{code}/cancel/{sizeID}", route.SaleDetailController.Cancel).Methods("DELETE")

	// POS (order barang masuk)
	authRouter.HandleFunc("/purchases", route.PurchaseController.Create).Methods("POST")
	authRouter.HandleFunc("/purchases", route.PurchaseController.List).Methods("GET")
	authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Cancel).Methods("DELETE")
	authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Get).Methods("GET")
	authRouter.HandleFunc("/purchases/{code}/cancel/{sizeID}", route.PurchaseDetailController.Cancel).Methods("DELETE")

	// laporan barang keluar [ list per barang ]
	authRouter.HandleFunc("/sales-and-product-reports/daily", route.SaleReportController.ListDaily).Methods("GET")

	// laporan keseluruhan barang keluar [ list barang terlaris ]
	authRouter.HandleFunc("/sales-and-product-reports/top-seller", route.SaleReportController.ListTopSeller).Methods("GET")

	// laporan keseluruhan barang keluar [ list barang terlaris per category ]
	authRouter.HandleFunc("/sales-and-product-reports/categories", route.SaleReportController.ListCategory).Methods("GET")

	// pengeluaran
	authRouter.HandleFunc("/expenses/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	authRouter.HandleFunc("/expenses", route.ExpenseController.Create).Methods("POST")
	authRouter.HandleFunc("/expenses", route.ExpenseController.List).Methods("GET")
	authRouter.HandleFunc("/expenses/{id}", route.ExpenseController.Update).Methods("PUT")
	authRouter.HandleFunc("/expenses/{id}", route.ExpenseController.Delete).Methods("DELETE")

	// pencatatan modal masuk/keluar
	// authRouter.HandleFunc("/capitals/consolidated", route.ExpenseController.ConsolidatedReport).Methods("GET")
	authRouter.HandleFunc("/capitals", route.CapitalController.Create).Methods("POST")
	authRouter.HandleFunc("/capitals", route.CapitalController.List).Methods("GET")
	authRouter.HandleFunc("/capitals/{id}", route.CapitalController.Update).Methods("PUT")
	authRouter.HandleFunc("/capitals/{id}", route.CapitalController.Delete).Methods("DELETE")

	// pencatatan penerimaan/pengeluaran uang
	authRouter.HandleFunc("/cash-bank-transactions", route.CashBankTransactionController.List).Methods("GET")

	// utang / piutang
	authRouter.HandleFunc("/debt", route.DebtController.List).Methods("GET")
	authRouter.HandleFunc("/debt/{id}", route.DebtController.Get).Methods("GET")
	authRouter.HandleFunc("/debt/{debtID}/payments", route.DebtPaymentController.Create).Methods("POST")
	authRouter.HandleFunc("/debt/{debtID}/payments/{id}", route.DebtPaymentController.Delete).Methods("DELETE")

	// ============================ skip =============================== //

	// ringkasan laporan keuangan [ owner only ]
	authRouter.HandleFunc("/finance-report/summary", route.FinanceController.GetSummary).Methods("GET")
	// laporan keuangan laba rugi [ owner, admin cabang ]
	authRouter.HandleFunc("/finance-report/profit-loss", route.FinanceController.GetProfitLoss).Methods("GET")
	// laporan keuangan arus kas [ owner, admin cabang ]
	authRouter.HandleFunc("/finance-report/cashflow", route.FinanceController.GetCashFlow).Methods("GET")
	// laporan keuangan neraca [ owner, admin cabang ]
	authRouter.HandleFunc("/finance-report/balance-sheet", route.FinanceController.GetBalanceSheet).Methods("GET")

	// ============================ skip =============================== //

	// stok barang
	authRouter.HandleFunc("/branch-inventory", route.BranchInventoryController.List).Methods("GET")

	// pergerakan stok barang masuk/keluar
	authRouter.HandleFunc("/inventory-movement", route.InventoryMovementController.List).Methods("GET")
	authRouter.HandleFunc("/inventory-movement", route.InventoryMovementController.Create).Methods("POST")
	// [owner only]
	authRouter.HandleFunc("/inventory-movement/summary", route.InventoryMovementController.Summary).Methods("GET")

}
