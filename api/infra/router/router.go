package router

import (
	"crypt-link/adapter/input/controller"
	"crypt-link/infra/middleware"
	"crypt-link/response"
	"errors"
	"net/http"
)

// Per-client-IP rate limit: sustained requests per second and burst size.
const (
	rateLimitRPS   = 2
	rateLimitBurst = 20
)

var errResourceNotFound = errors.New("resource not found")

// Setup registers the API routes and returns the HTTP handler, with per-client-IP
// rate limiting for deployments where the API sees the client address.
func Setup(msgCtrl *controller.MessageController) http.Handler {
	limiter := middleware.NewRateLimiter(rateLimitRPS, rateLimitBurst)
	return middleware.SecurityHeaders(limiter.Middleware(routes(msgCtrl)))
}

// SetupWithoutRateLimit returns the API handler without the in-memory rate
// limiter, for runtimes where in-memory state does not persist between
// requests (Cloudflare Workers creates a new instance per request) and rate
// limiting is enforced in front of the API.
func SetupWithoutRateLimit(msgCtrl *controller.MessageController) http.Handler {
	return middleware.SecurityHeaders(routes(msgCtrl))
}

func routes(msgCtrl *controller.MessageController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/message", msgCtrl.CreateMessage)
	mux.HandleFunc("GET /api/message/{id}", msgCtrl.GetMessage)

	// Fallback for any unmatched route.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, errResourceNotFound)
	})
	return mux
}
