package routers

import (
	"crypt-link/adapter/input/controller"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Route struct {
	URI      string
	Method   string
	Function func(http.ResponseWriter, *http.Request)
}

func Configurate(r *mux.Router, msgCtrl *controller.MessageController) *mux.Router {
	routes := BuildMessageRoutes(msgCtrl)

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
