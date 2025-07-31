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
	branchRepository := repository.NewBranchRepository(config.Log)
	userRepository := repository.NewUserRepository(config.Log)

	// setup use cases
	branchUseCase := usecase.NewBranchUseCase(config.DB, config.Log, config.Validate, branchRepository)
	userUseCase := usecase.NewUserUseCase(config.DB, config.Log, config.Validate, userRepository)

	// setup controller
	userController := http.NewUserController(userUseCase, config.Log)
	branchController := http.NewBranchController(branchUseCase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUseCase)

	routeConfig := route.RouteConfig{
		Router:           config.Router,
		AuthMiddleware:   authMiddleware,
		UserController:   userController,
		BranchController: branchController,
	}
	routeConfig.Setup()

}
