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
	UserController   *http.UserController
}

func (route *RouteConfig) Setup() {
	route.SetupGuestRoute()
	route.SetupAuthRoute()
}

func (route *RouteConfig) SetupGuestRoute() {
	// routes that do not require authentication
	route.Router.HandleFunc("/register", route.UserController.Register).Methods("POST")
	route.Router.HandleFunc("/login", route.UserController.Login).Methods("POST")
}

func (route *RouteConfig) SetupAuthRoute() {

	// Buat subrouter khusus untuk route yang butuh auth
	authRouter := route.Router.PathPrefix("/").Subrouter()
	authRouter.Use(route.AuthMiddleware)

	authRouter.HandleFunc("/logout", route.UserController.Logout).Methods("POST")

	authRouter = route.Router.PathPrefix("/api/v1/").Subrouter()
	authRouter.Use(route.AuthMiddleware)

	// route master data
	authRouter.HandleFunc("/branchs", route.BranchController.Create).Methods("POST")
	authRouter.HandleFunc("/branchs", route.BranchController.List).Methods("GET")
	authRouter.HandleFunc("/branchs/{id}", route.BranchController.Update).Methods("PUT")
	authRouter.HandleFunc("/branchs/{id}", route.BranchController.Delete).Methods("DELETE")
	// end route master data

	authRouter.HandleFunc("/users", route.UserController.Register).Methods("POST")
	authRouter.HandleFunc("/users", route.UserController.List).Methods("GET")
	authRouter.HandleFunc("/users/{username}", route.UserController.Update).Methods("PUT")
	authRouter.HandleFunc("/users/{username}", route.UserController.Delete).Methods("DELETE")

}
