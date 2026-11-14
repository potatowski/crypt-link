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

// routes dispatches the API requests. It matches paths and methods explicitly
// instead of using Go 1.22 ServeMux patterns ("POST /api/message",
// "GET /api/message/{id}"), which TinyGo's ServeMux does not support, while
// keeping their behaviour: segments are compared after unescaping, GET routes
// also match HEAD (the controller rejects it), {id} is one non-empty segment,
// and anything else gets a 404.
func routes(msgCtrl *controller.MessageController) http.Handler {
	notFound := func(w http.ResponseWriter) {
		response.Error(w, http.StatusNotFound, errResourceNotFound)
	}

	mux := http.NewServeMux() // still cleans paths and redirects like before
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		segments, ok := pathSegments(r.URL.EscapedPath())
		switch {
		case !ok:
			notFound(w)
		case len(segments) == 2 && isMessagePath(segments) && r.Method == http.MethodPost:
			msgCtrl.CreateMessage(w, r)
		case len(segments) == 3 && isMessagePath(segments) && segments[2] != "" &&
			(r.Method == http.MethodGet || r.Method == http.MethodHead):
			r.SetPathValue("id", segments[2])
			msgCtrl.GetMessage(w, r)
		default:
			notFound(w)
		}
	})
	return mux
}

func isMessagePath(segments []string) bool {
	return segments[0] == "api" && segments[1] == "message"
}

// pathSegments splits an escaped path into unescaped segments.
func pathSegments(escapedPath string) ([]string, bool) {
	segments := strings.Split(strings.TrimPrefix(escapedPath, "/"), "/")
	for i, seg := range segments {
		unescaped, err := url.PathUnescape(seg)
		if err != nil {
			return nil, false
		}
		segments[i] = unescaped
	}
	return segments, true
}
