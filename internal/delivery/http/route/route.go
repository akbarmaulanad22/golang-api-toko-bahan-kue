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

	ProductController  *http.ProductController
	UserController     *http.UserController
	SizeController     *http.SizeController
	SaleController     *http.SaleController
	PurchaseController *http.PurchaseController
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

	authRouter.HandleFunc("/gate/auth/logout", route.UserController.Logout).Methods("POST")
	authRouter.HandleFunc("/gate/auth/me", route.UserController.Current).Methods("GET")

	authRouter = route.Router.PathPrefix("/api/v1/").Subrouter()
	authRouter.Use(route.AuthMiddleware)

	// master data
	authRouter.HandleFunc("/branches", route.BranchController.Create).Methods("POST")
	authRouter.HandleFunc("/branches", route.BranchController.List).Methods("GET")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Update).Methods("PUT")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/branches/{id}", route.BranchController.Get).Methods("GET")

	authRouter.HandleFunc("/roles", route.RoleController.Create).Methods("POST")
	authRouter.HandleFunc("/roles", route.RoleController.List).Methods("GET")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Update).Methods("PUT")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/roles/{id}", route.RoleController.Get).Methods("GET")

	authRouter.HandleFunc("/categories", route.CategoryController.Create).Methods("POST")
	authRouter.HandleFunc("/categories", route.CategoryController.List).Methods("GET")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Update).Methods("PUT")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/categories/{id}", route.CategoryController.Get).Methods("GET")

	authRouter.HandleFunc("/distributors", route.DistributorController.Create).Methods("POST")
	authRouter.HandleFunc("/distributors", route.DistributorController.List).Methods("GET")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Update).Methods("PUT")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/distributors/{id}", route.DistributorController.Get).Methods("GET")
	// end master data

	authRouter.HandleFunc("/products", route.ProductController.Create).Methods("POST")
	authRouter.HandleFunc("/products", route.ProductController.List).Methods("GET")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Update).Methods("PUT")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/products/{sku}", route.ProductController.Get).Methods("GET")

	authRouter.HandleFunc("/users", route.UserController.Register).Methods("POST")
	authRouter.HandleFunc("/users", route.UserController.List).Methods("GET")
	authRouter.HandleFunc("/users/{username}", route.UserController.Update).Methods("PUT")
	authRouter.HandleFunc("/users/{username}", route.UserController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/users/{username}", route.UserController.Get).Methods("GET")

	authRouter.HandleFunc("/products/{productSKU}/sizes", route.SizeController.Create).Methods("POST")
	authRouter.HandleFunc("/products/{productSKU}/sizes", route.SizeController.List).Methods("GET")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Update).Methods("PUT")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/products/{productSKU}/sizes/{id}", route.SizeController.Get).Methods("GET")

	authRouter.HandleFunc("/sales", route.SaleController.Create).Methods("POST")
	authRouter.HandleFunc("/sales", route.SaleController.List).Methods("GET")
	authRouter.HandleFunc("/sales/{code}", route.SaleController.Update).Methods("PUT")
	authRouter.HandleFunc("/sales/{code}", route.SaleController.Get).Methods("GET")

	authRouter.HandleFunc("/purchases", route.PurchaseController.Create).Methods("POST")
	authRouter.HandleFunc("/purchases", route.PurchaseController.List).Methods("GET")
	authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Update).Methods("PUT")
	authRouter.HandleFunc("/purchases/{code}", route.PurchaseController.Get).Methods("GET")

	authRouter.HandleFunc("/sales-report", route.SaleController.ListReport).Methods("GET")
	// authRouter.HandleFunc("/sales-report/{branchID}/daily", route.SaleSummaryController.ListDailySaleSummary).Methods("GET")
}
