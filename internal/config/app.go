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

	// master data usecase
	branchUseCase := usecase.NewBranchUseCase(config.DB, config.Log, config.Validate, branchRepository)
	roleUseCase := usecase.NewRoleUseCase(config.DB, config.Log, config.Validate, roleRepository)
	categoryUseCase := usecase.NewCategoryUseCase(config.DB, config.Log, config.Validate, categoryRepository)
	distributorUseCase := usecase.NewDistributorUseCase(config.DB, config.Log, config.Validate, distributorRepository)

	// data usecase
	productUseCase := usecase.NewProductUseCase(config.DB, config.Log, config.Validate, productRepository)
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)

	// master data controller
	branchController := http.NewBranchController(branchUseCase, config.Log)
	roleController := http.NewRoleController(roleUseCase, config.Log)
	categoryController := http.NewCategoryController(categoryUseCase, config.Log)
	distributorController := http.NewDistributorController(distributorUseCase, config.Log)

	// data controller
	userController := http.NewUserController(userUseCase, config.Log)
	productController := http.NewProductController(productUseCase, config.Log)

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
		ProductController: productController,
		UserController:    userController,
	}
	routeConfig.Setup()

}
