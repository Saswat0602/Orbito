package middleware

import (
	"net/http"
	"strings"
)

// CORS handles Cross-Origin Resource Sharing mirroring NestJS configuration.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originMap := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		originMap[strings.TrimSpace(origin)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check origin against whitelist
			if origin != "" && (originMap["*"] || originMap[origin]) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-correlation-id, x-request-id, Accept, Origin")
				w.Header().Set("Access-Control-Expose-Headers", "x-correlation-id, x-request-id, Content-Disposition")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
