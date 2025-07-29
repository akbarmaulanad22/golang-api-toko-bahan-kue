package main

import (
	"tokobahankue/internal/config"

	"github.com/gorilla/mux"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	router := mux.NewRouter()
	app := config.NewServer(viperConfig, router)

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		Router:   router,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
	})

	err := app.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
