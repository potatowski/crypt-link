package router

import (
	"crypt-link/adapter/input/controller"
	"crypt-link/infra/router/routers"

	"github.com/gorilla/mux"
)

func Setup(msgCtrl *controller.MessageController) *mux.Router {
	r := mux.NewRouter()
	return routers.Configurate(r, msgCtrl)
}
