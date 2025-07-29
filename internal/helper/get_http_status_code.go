package helper

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func GetStatusCode(err error) int {
	// Contoh pengecekan error berdasarkan string atau tipe error khusus

	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}

	if errors.Is(err, context.Canceled) {
		return http.StatusBadRequest
	}

	// Contoh error validation
	if strings.Contains(err.Error(), "bad request") {
		return http.StatusBadRequest
	}

	// Contoh error duplicate email
	if strings.Contains(err.Error(), "conflict") {
		return http.StatusConflict
	}

	// Contoh error unauthorized
	if strings.Contains(err.Error(), "unauthorized") {
		return http.StatusUnauthorized
	}

	// Contoh error not found
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}

	// Default fallback
	return http.StatusInternalServerError
}
