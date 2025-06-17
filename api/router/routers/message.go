package routers

import (
	"crypt-link/controller"
	"net/http"
)

var endpointsMessage = []Route{
	{
		URI:      "/api/message",
		Method:   http.MethodPost,
		Function: controller.CreateMessage,
	},
	{
		URI:      "/api/message/{id}",
		Method:   http.MethodGet,
		Function: controller.GetMessage,
	},
}
