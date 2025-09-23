package config

import (
	"tokobahankue/internal/delivery/http"
	"tokobahankue/internal/delivery/http/middleware"
	"tokobahankue/internal/delivery/http/route"
	"tokobahankue/internal/repository"
	"tokobahankue/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	Router   *mux.Router
	DB       *gorm.DB
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(config *BootstrapConfig) {

	// master data repository
	branchRepository := repository.NewBranchRepository(config.Log)
	roleRepository := repository.NewRoleRepository(config.Log)
	categoryRepository := repository.NewCategoryRepository(config.Log)
	distributorRepository := repository.NewDistributorRepository(config.Log)

	// data repository
	dashboardRepository := repository.NewDashboardRepository(config.Log)
	productRepository := repository.NewProductRepository(config.Log)
	userRepository := repository.NewUserRepository(config.Log)
	sizeRepository := repository.NewSizeRepository(config.Log)
	saleRepository := repository.NewSaleRepository(config.Log)
	purchaseRepository := repository.NewPurchaseRepository(config.Log)
	saleReportRepository := repository.NewSaleReportRepository(config.Log)
	expenseRepository := repository.NewExpenseRepository(config.Log)
	capitalRepository := repository.NewCapitalRepository(config.Log)
	cashBankTransactionRepository := repository.NewCashBankTransactionRepository(config.Log)
	financeRepository := repository.NewFinanceRepository(config.Log)
	debtRepository := repository.NewDebtRepository(config.Log)
	branchInventoryRepository := repository.NewBranchInventoryRepository(config.Log)
	inventoryMovementRepository := repository.NewInventoryMovementRepository(config.Log)
	debtPaymentRepository := repository.NewDebtPaymentRepository(config.Log)
	saleDetailRepository := repository.NewSaleDetailRepository(config.Log)
	salePaymentRepository := repository.NewSalePaymentRepository(config.Log)
	purchaseDetailRepository := repository.NewPurchaseDetailRepository(config.Log)
	purchasePaymentRepository := repository.NewPurchasePaymentRepository(config.Log)
	purchaseReportRepository := repository.NewPurchaseReportRepository(config.Log)

	// master data usecase
	branchUseCase := usecase.NewBranchUseCase(config.DB, config.Log, config.Validate, branchRepository)
	roleUseCase := usecase.NewRoleUseCase(config.DB, config.Log, config.Validate, roleRepository)
	categoryUseCase := usecase.NewCategoryUseCase(config.DB, config.Log, config.Validate, categoryRepository)
	distributorUseCase := usecase.NewDistributorUseCase(config.DB, config.Log, config.Validate, distributorRepository)

	// data usecase
	dashboardUseCase := usecase.NewDashboardUseCase(config.DB, config.Log, config.Validate, dashboardRepository)
	productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)
	sizeUseCase := usecase.NewSizeUseCase(config.DB, config.Log, config.Validate, sizeRepository)
	saleUseCase := usecase.NewSaleUseCase(config.DB, config.Log, config.Validate, saleRepository, saleDetailRepository, salePaymentRepository, debtRepository, debtPaymentRepository, sizeRepository, cashBankTransactionRepository, branchInventoryRepository, inventoryMovementRepository)
	purchaseUseCase := usecase.NewPurchaseUseCase(config.DB, config.Log, config.Validate, purchaseRepository, purchaseDetailRepository, purchasePaymentRepository, debtRepository, debtPaymentRepository, sizeRepository, cashBankTransactionRepository, branchInventoryRepository, inventoryMovementRepository)
	saleReportUseCase := usecase.NewSaleReportUseCase(config.DB, config.Log, config.Validate, saleReportRepository)
	expenseUseCase := usecase.NewExpenseUseCase(config.DB, config.Log, config.Validate, expenseRepository, cashBankTransactionRepository)
	capitalUseCase := usecase.NewCapitalUseCase(config.DB, config.Log, config.Validate, capitalRepository, cashBankTransactionRepository)
	cashBankTransactionUseCase := usecase.NewCashBankTransactionUseCase(config.DB, config.Log, config.Validate, cashBankTransactionRepository)
	financeUseCase := usecase.NewFinanceUseCase(config.DB, config.Log, config.Validate, financeRepository)
	debtUseCase := usecase.NewDebtUseCase(config.DB, config.Log, config.Validate, debtRepository)
	branchInventoryUseCase := usecase.NewBranchInventoryUseCase(config.DB, config.Log, config.Validate, branchInventoryRepository)
	inventoryMovementUseCase := usecase.NewInventoryMovementUseCase(config.DB, config.Log, config.Validate, inventoryMovementRepository, branchInventoryRepository)
	debtPaymentUseCase := usecase.NewDebtPaymentUseCase(config.DB, config.Log, config.Validate, debtPaymentRepository, cashBankTransactionRepository, debtRepository)
	saleDetailUseCase := usecase.NewSaleDetailUseCase(config.DB, config.Log, config.Validate, saleDetailRepository, saleRepository)
	purchaseDetailUseCase := usecase.NewPurchaseDetailUseCase(config.DB, config.Log, config.Validate, purchaseDetailRepository, purchaseRepository)
	purchaseReportUseCase := usecase.NewPurchaseReportUseCase(config.DB, config.Log, config.Validate, purchaseReportRepository)

	// master data controller
	branchController := http.NewBranchController(branchUseCase, config.Log)
	roleController := http.NewRoleController(roleUseCase, config.Log)
	categoryController := http.NewCategoryController(categoryUseCase, config.Log)
	distributorController := http.NewDistributorController(distributorUseCase, config.Log)

	// data controller
	dashboardController := http.NewDashboardController(dashboardUseCase, config.Log)
	productController := http.NewProductController(productUseCase, config.Log)
	userController := http.NewUserController(userUseCase, config.Log)
	sizeController := http.NewSizeController(sizeUseCase, config.Log)
	saleController := http.NewSaleController(saleUseCase, config.Log)
	purchaseController := http.NewPurchaseController(purchaseUseCase, config.Log)
	saleReportController := http.NewSaleReportController(saleReportUseCase, config.Log)
	expenseController := http.NewExpenseController(expenseUseCase, config.Log)
	capitalController := http.NewCapitalController(capitalUseCase, config.Log)
	cashBankTransactionController := http.NewCashBankTransactionController(cashBankTransactionUseCase, config.Log)
	financeController := http.NewFinanceController(financeUseCase, config.Log)
	debtController := http.NewDebtController(debtUseCase, config.Log)
	branchInventoryController := http.NewBranchInventoryController(branchInventoryUseCase, config.Log)
	inventoryMovementController := http.NewInventoryMovementController(inventoryMovementUseCase, config.Log)
	debtPaymentController := http.NewDebtPaymentController(debtPaymentUseCase, config.Log)
	saleDetailController := http.NewSaleDetailController(saleDetailUseCase, config.Log)
	purchaseDetailController := http.NewPurchaseDetailController(purchaseDetailUseCase, config.Log)
	purchaseReportController := http.NewPurchaseReportController(purchaseReportUseCase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		Router:         config.Router,
		AuthMiddleware: authMiddleware,
		// master data controller
		BranchController:      branchController,
		RoleController:        roleController,
		CategoryController:    categoryController,
		DistributorController: distributorController,
		// data controller
		DashboardController:           dashboardController,
		ProductController:             productController,
		UserController:                userController,
		SizeController:                sizeController,
		SaleController:                saleController,
		PurchaseController:            purchaseController,
		SaleReportController:          saleReportController,
		ExpenseController:             expenseController,
		CapitalController:             capitalController,
		CashBankTransactionController: cashBankTransactionController,
		FinanceController:             financeController,
		DebtController:                debtController,
		BranchInventoryController:     branchInventoryController,
		InventoryMovementController:   inventoryMovementController,
		DebtPaymentController:         debtPaymentController,
		SaleDetailController:          saleDetailController,
		PurchaseDetailController:      purchaseDetailController,
		PurchaseReportController:      purchaseReportController,
	}
	routeConfig.Setup()

}
