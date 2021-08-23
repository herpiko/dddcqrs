package middleware

import (
	"net/http"
)

// Global middleware
func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(r.Context()))
	})
}
