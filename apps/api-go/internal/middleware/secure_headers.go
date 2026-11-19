package middleware

import (
	"net/http"
)

// SecureHeaders injects Helmet-equivalent security headers into HTTP responses.
func SecureHeaders(isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-XSS-Protection", "0")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("X-Download-Options", "noopen")
			w.Header().Set("X-DNS-Prefetch-Control", "off")

			if isProduction {
				w.Header().Set("Strict-Transport-Security", "max-age=15552000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}
