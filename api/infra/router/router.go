package router

import (
	"crypt-link/adapter/input/controller"
	"crypt-link/infra/middleware"
	"crypt-link/response"
	"errors"
	"net/http"
	"net/url"
	"strings"
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

const (
	messagePath   = "/api/message"
	messagePrefix = messagePath + "/"
)

// routes dispatches the API requests. It matches paths and methods explicitly
// instead of using Go 1.22 ServeMux patterns ("POST /api/message",
// "GET /api/message/{id}"), which TinyGo's ServeMux does not support, while
// keeping their behaviour: GET routes also match HEAD (the controller rejects
// it), {id} is one non-empty path segment, and anything else gets a 404.
func routes(msgCtrl *controller.MessageController) http.Handler {
	notFound := func(w http.ResponseWriter) {
		response.Error(w, http.StatusNotFound, errResourceNotFound)
	}

	mux := http.NewServeMux() // still cleans paths and redirects like before
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.EscapedPath()
		switch {
		case path == messagePath && r.Method == http.MethodPost:
			msgCtrl.CreateMessage(w, r)
		case strings.HasPrefix(path, messagePrefix) && (r.Method == http.MethodGet || r.Method == http.MethodHead):
			segment := strings.TrimPrefix(path, messagePrefix)
			id, err := url.PathUnescape(segment)
			if segment == "" || strings.Contains(segment, "/") || err != nil {
				notFound(w)
				return
			}
			r.SetPathValue("id", id)
			msgCtrl.GetMessage(w, r)
		default:
			notFound(w)
		}
	})
	return mux
}
