package router

import (
	"crypt-link/controller"
	"crypt-link/database"
	"crypt-link/repository"
	"crypt-link/service"

	"net/http"

	"github.com/gorilla/mux"
)

// Setup initializes the application's HTTP router with all necessary routes and handlers.
// It establishes a MongoDB connection using the provided configuration, sets up the repository,
// service, and controller layers for message handling, and registers API endpoints for creating
// and retrieving messages. Additionally, it serves static files from the "./web" directory for
// all other routes.
//
// Returns:
//   - *mux.Router: The configured HTTP router ready to be used by the server.
func Setup() *mux.Router {
	client, cancelFunc, _ := database.Connect()
	defer cancelFunc()

	repo := repository.NewMessageRepository(client)
	svc := service.NewMessageService(repo)
	ctrl := controller.NewMessageController(svc)

	r := mux.NewRouter()
	r.HandleFunc("/api/message", ctrl.CreateMessage).Methods("POST")
	r.HandleFunc("/api/message/{id}", ctrl.GetMessage).Methods("GET")

	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/").Handler(fs)

	return r
}
