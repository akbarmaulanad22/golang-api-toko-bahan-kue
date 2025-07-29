package config

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

func NewServer(viper *viper.Viper, router *mux.Router) *http.Server {

	port := viper.GetString("SERVER_PORT")

	// Konfigurasi CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Ganti dengan origin Flutter Anda
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	// Gunakan middleware CORS
	handler := corsHandler.Handler(router)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server

}
