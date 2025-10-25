package middleware

import (
	"context"
	"net/http"
	"tokobahankue/internal/model"
	"tokobahankue/internal/usecase"

	"github.com/gorilla/mux"
)

func NewAuth(userUseCase *usecase.UserUseCase) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				token = "NOT_FOUND"
			}
			userUseCase.Log.Debugf("Authorization : %s", token)

			request := &model.VerifyUserRequest{Token: token}
			auth, err := userUseCase.Verify(r.Context(), request)
			if err != nil {
				userUseCase.Log.Warnf("Failed find user by token : %+v", err)

				appErr := &model.AppError{
					Message: "unauthorized",
				}
				ctx := context.WithValue(r.Context(), "middleware_error", appErr)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			userUseCase.Log.Debugf("User : %+v", auth.Username)
			ctx := context.WithValue(r.Context(), "auth", auth)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(r *http.Request) *model.Auth {
	auth, _ := r.Context().Value("auth").(*model.Auth)
	return auth
}
