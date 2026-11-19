package middleware

import (
	"net/http"
)

// BodyLimit limits the maximum size of incoming request bodies.
// Default in Orbito is 10MB (10485760 bytes).
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = 10 << 20 // 10 MB default
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}
