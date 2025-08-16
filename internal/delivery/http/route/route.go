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

	DashboardController       *http.DashboardController
	ProductController         *http.ProductController
	UserController            *http.UserController
	SizeController            *http.SizeController
	SaleController            *http.SaleController
	PurchaseController        *http.PurchaseController
	FinancialReportController *http.FinancialReportController
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

	// POS (order barang masuk)
	// authRouter.HandleFunc("/purchases", route.PurchaseController.Create).Methods("POST")
	// authRouter.HandleFunc("/purchases", route.PurchaseController.List).Methods("GET")
	// authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Update).Methods("PUT")
	// authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Get).Methods("GET")

	// laporan barang keluar [ list per barang ]
	authRouter.HandleFunc("/sales-and-product-reports/sales", route.SaleController.ListReport).Methods("GET")

	// laporan keseluruhan barang keluar per cabang
	authRouter.HandleFunc("/sales-and-product-reports/branch-sales", route.SaleController.ListBranchSaleReport).Methods("GET")

	// laporan keseluruhan barang terlaris [ bisa filter berdasarkan cabang ]
	authRouter.HandleFunc("/sales-and-product-reports/best-selling-product", route.SaleController.ListBestSellingProduct).Methods("GET")

	// laporan keuangan
	// authRouter.HandleFunc("/financial-reports", route.FinancialReportController.List).Methods("GET")

}
