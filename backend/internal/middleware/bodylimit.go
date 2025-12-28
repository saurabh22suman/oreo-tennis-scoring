package middleware

import (
	"net/http"
)

// MaxBodySize is the default maximum request body size (1MB).
const MaxBodySize = 1 << 20 // 1MB

// LimitBody returns middleware that limits request body size.
// This prevents DoS attacks via large payloads.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for requests without body
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Wrap body with size limiter
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
