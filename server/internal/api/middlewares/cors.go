package middlewares

import (
	"net/http"
)

// CORS middleware handles cross-origin resource sharing headers and preflight OPTIONS requests.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// If allowedOrigin is set, check match, or default to frontend URL / origin
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			} else if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, Origin")
			w.Header().Set("Access-Control-Expose-Headers", "Link, Content-Length, Access-Control-Allow-Origin")
			w.Header().Set("Access-Control-Max-Age", "300")

			// Handle preflight OPTIONS requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
