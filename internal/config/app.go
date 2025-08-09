package config

import (
	// "tokobahankue/internal/delivery/http"
	// "tokobahankue/internal/delivery/http/middleware"
	// "tokobahankue/internal/delivery/http/route"
	// "tokobahankue/internal/repository"
	// "tokobahankue/internal/usecase"

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
	productRepository := repository.NewProductRepository(config.Log)
	userRepository := repository.NewUserRepository(config.Log)
	sizeRepository := repository.NewSizeRepository(config.Log)
	saleRepository := repository.NewSaleRepository(config.Log)
	purchaseRepository := repository.NewPurchaseRepository(config.Log)

	// master data usecase
	branchUseCase := usecase.NewBranchUseCase(config.DB, config.Log, config.Validate, branchRepository)
	roleUseCase := usecase.NewRoleUseCase(config.DB, config.Log, config.Validate, roleRepository)
	categoryUseCase := usecase.NewCategoryUseCase(config.DB, config.Log, config.Validate, categoryRepository)
	distributorUseCase := usecase.NewDistributorUseCase(config.DB, config.Log, config.Validate, distributorRepository)

	// data usecase
	productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)
	sizeUseCase := usecase.NewSizeUseCase(config.DB, config.Log, config.Validate, sizeRepository)
	saleUseCase := usecase.NewSaleUseCase(config.DB, config.Log, config.Validate, saleRepository)
	purchaseUseCase := usecase.NewPurchaseUseCase(config.DB, config.Log, config.Validate, purchaseRepository)

	// master data controller
	branchController := http.NewBranchController(branchUseCase, config.Log)
	roleController := http.NewRoleController(roleUseCase, config.Log)
	categoryController := http.NewCategoryController(categoryUseCase, config.Log)
	distributorController := http.NewDistributorController(distributorUseCase, config.Log)

	// data controller
	productController := http.NewProductController(productUseCase, config.Log)
	userController := http.NewUserController(userUseCase, config.Log)
	sizeController := http.NewSizeController(sizeUseCase, config.Log)
	saleController := http.NewSaleController(saleUseCase, config.Log)
	purchaseController := http.NewPurchaseController(purchaseUseCase, config.Log)

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
		ProductController:  productController,
		UserController:     userController,
		SizeController:     sizeController,
		SaleController:     saleController,
		PurchaseController: purchaseController,
	}
	routeConfig.Setup()

}
