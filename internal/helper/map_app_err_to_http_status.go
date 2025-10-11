package helper

import (
	"net/http"
	"strings"
	"tokobahankue/internal/model"
)

func MapAppErrorToHTTPStatus(err *model.AppError) int {
	msg := strings.ToLower(err.Message)

	switch {
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.Contains(msg, "invalid"):
		return http.StatusBadRequest
	case strings.Contains(msg, "validation"):
		return http.StatusUnprocessableEntity
	case strings.Contains(msg, "unauthorized"):
		return http.StatusUnauthorized
	case strings.Contains(msg, "forbidden"):
		return http.StatusForbidden
	case strings.Contains(msg, "conflict"), strings.Contains(msg, "duplicate"):
		return http.StatusConflict
	case strings.Contains(msg, "timeout"):
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
