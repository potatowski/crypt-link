package router

import (
	"crypt-link/router/routers"

	"github.com/gorilla/mux"
)

// Setup initializes a new Gorilla Mux router and applies additional configuration
// using the routers.Configurate function. It returns the configured router instance.
//
// Returns:
//   - *mux.Router: The configured HTTP router ready to be used by the server.
func Setup() *mux.Router {
	r := mux.NewRouter()
	return routers.Configurate(r)
}
