package middleware

import (
	"errors"
	"net/http"
	"tokobahankue/internal/helper"
	"tokobahankue/internal/model"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func WithErrorHandler(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if errVal := r.Context().Value("middleware_error"); errVal != nil {
			if appErr, ok := errVal.(*model.AppError); ok {
				status := helper.MapAppErrorToHTTPStatus(appErr)
				helper.WriteJSON(w, status, map[string]any{
					"error": appErr,
				})
				return
			}
		}

		if err := h(w, r); err != nil {
			var appErr *model.AppError
			if errors.As(err, &appErr) {
				status := helper.MapAppErrorToHTTPStatus(appErr)
				helper.WriteJSON(w, status, map[string]any{
					"error": appErr,
				})
				return
			}

			helper.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"error": map[string]any{
					"message": "internal server error",
				},
			})
		}
	}
}
