package routers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Route represents an HTTP route with its URI, HTTP method, and handler function.
// URI specifies the endpoint path.
// Method specifies the HTTP method (e.g., "GET", "POST").
// Function is the handler to be executed when the route is matched.
type Route struct {
	URI      string
	Method   string
	Function func(http.ResponseWriter, *http.Request)
}

// Configurate sets up the provided mux.Router by registering all routes defined in the
// endpointsMessage slice. Each route is associated with its URI, HTTP method, and handler function.
//
// Returns:
//   - *mux.Router: The configured HTTP router ready to be used by the server.
func Configurate(r *mux.Router) *mux.Router {
	routes := endpointsMessage

	for _, route := range routes {
		fmt.Println("Registering route:", route.URI, "with method:", route.Method)

		r.HandleFunc(route.URI, route.Function).Methods(route.Method)
	}

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "resource not found"}`))
	})

	return r
}
