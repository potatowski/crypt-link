package middleware

import "net/http"

// EdgeHeaders adds the API response headers that nginx sets in the Docker
// deployment, for runtimes without nginx in front of the API (Cloudflare Workers).
func EdgeHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Strict-Transport-Security", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
