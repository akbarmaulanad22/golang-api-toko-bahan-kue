package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type MiddlewareCustom func(roles ...string) mux.MiddlewareFunc

func NewRolesMiddleware(allowedRoles ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := GetUser(r)

			log.Print(auth)

			if auth == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			for _, role := range allowedRoles {
				if strings.EqualFold(auth.Role, role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
		})
	}
}

// func requireRoles(allowedRoles ...string) mux.MiddlewareFunc {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			auth := GetUser(r)
// 			if auth == nil {
// 				http.Error(w, "unauthorized", http.StatusUnauthorized)
// 				return
// 			}

// 			for _, role := range allowedRoles {
// 				if strings.EqualFold(auth.Role, role) {
// 					next.ServeHTTP(w, r)
// 					return
// 				}
// 			}

// 			http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
// 		})
// 	}
// }

// func NewAdminMiddleware() mux.MiddlewareFunc {
// 	return requireRoles("ADMIN")
// }

// func NewOwnerMiddleware() mux.MiddlewareFunc {
// 	return requireRoles("OWNER")
// }
