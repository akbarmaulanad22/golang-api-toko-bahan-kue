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

	// setup repositories

	// master data repository
	branchRepository := repository.NewBranchRepository(config.Log)
	roleRepository := repository.NewRoleRepository(config.Log)

	// data repo
	userRepository := repository.NewUserRepository(config.Log)

	// end setup repository

	// setup use cases

	// master data usecase
	branchUseCase := usecase.NewBranchUseCase(config.DB, config.Log, config.Validate, branchRepository)
	roleUseCase := usecase.NewRoleUseCase(config.DB, config.Log, config.Validate, roleRepository)

	// data usecase
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)

	// setup controller

	// master data controller
	branchController := http.NewBranchController(branchUseCase, config.Log)
	roleController := http.NewRoleController(roleUseCase, config.Log)

	// master controller
	userController := http.NewUserController(userUseCase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		Router:         config.Router,
		AuthMiddleware: authMiddleware,

		// master data controller
		BranchController: branchController,
		RoleController:   roleController,

		//
		UserController: userController,
	}
	routeConfig.Setup()

}
