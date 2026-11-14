package middleware

import "net/http"

// SecurityHeaders sets headers that keep API responses (which may carry
// one-time secrets) out of caches and prevent MIME sniffing.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Cache-Control", "no-store")
		h.Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}
