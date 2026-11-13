package router

import (
	"crypt-link/adapter/input/controller"
	"crypt-link/response"
	"errors"
	"net/http"
)

var errResourceNotFound = errors.New("resource not found")

// Setup registers the API routes and returns the HTTP handler.
func Setup(msgCtrl *controller.MessageController) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/message", msgCtrl.CreateMessage)
	mux.HandleFunc("GET /api/message/{id}", msgCtrl.GetMessage)

	// Fallback for any unmatched route.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, errResourceNotFound)
	})

	return mux
}
