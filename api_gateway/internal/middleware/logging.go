package middleware

import (
	"common_library/logging"
	"net/http"
)

func NewLoggingMiddleware(logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(logging.ContextWithLogger(r.Context(), logger))
			next.ServeHTTP(w, r)
		})
	}
}
