package router

import (
	"context"
	"crypt-link/config"
	"crypt-link/controller"
	"crypt-link/repository"
	"crypt-link/service"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Setup(cfg config.Config) *mux.Router {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))

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
