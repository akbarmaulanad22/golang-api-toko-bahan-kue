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

	/*
		master data
		...
	*/
	BranchController *http.BranchController
	RoleController   *http.RoleController

	UserController *http.UserController
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

	authRouter.HandleFunc("/logout", route.UserController.Logout).Methods("POST")

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
	// end master data

	authRouter.HandleFunc("/users", route.UserController.Register).Methods("POST")
	authRouter.HandleFunc("/users", route.UserController.List).Methods("GET")
	authRouter.HandleFunc("/users/{username}", route.UserController.Update).Methods("PUT")
	authRouter.HandleFunc("/users/{username}", route.UserController.Delete).Methods("DELETE")
	authRouter.HandleFunc("/users/{username}", route.UserController.Get).Methods("GET")

}
