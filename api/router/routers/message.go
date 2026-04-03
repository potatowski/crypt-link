package routers

import (
	"crypt-link/adapter/input/controller"
	"net/http"
)

func BuildMessageRoutes(messageController *controller.MessageController) []Route {
	return []Route{
		{
			URI:      "/api/message",
			Method:   http.MethodPost,
			Function: messageController.CreateMessage,
		},
		{
			URI:      "/api/message/{id}",
			Method:   http.MethodGet,
			Function: messageController.GetMessage,
		},
	}
}
