package routers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Route is a struct to representation of all api routes
type Route struct {
	URI      string
	Method   string
	Function func(http.ResponseWriter, *http.Request)
}

// Configurate added all routes in mux router
func Configurate(r *mux.Router) *mux.Router {
	routes := endpointsMessage

	for _, route := range routes {
		fmt.Println("Registering route:", route.URI, "with method:", route.Method)

		r.HandleFunc(route.URI, route.Function).Methods(route.Method)
	}

	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/").Handler(fs)

	return r
}
