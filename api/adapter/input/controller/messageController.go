package controller

import (
	"crypt-link/core/port"
	"crypt-link/response"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type createRequest struct {
	ID        string `json:"id"`
	Encrypted string `json:"encrypted"`
}

const (
	ValidityTime = time.Hour * 24 * 30
)

type MessageController struct {
	service port.MessageService
}

func NewMessageController(service port.MessageService) *MessageController {
	return &MessageController{
		service: service,
	}
}

func (c *MessageController) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	expiresAt := time.Now().Add(time.Duration(ValidityTime))

	if err := c.service.Create(req.ID, req.Encrypted, expiresAt); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Message created successfully"})
}

func (c *MessageController) GetMessage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		response.Error(w, http.StatusBadRequest, errors.New("ID is required"))
		return
	}

	msg, err := c.service.GetAndInvalidate(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, errors.New("message not found or expired"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"encrypted": msg.Encrypted})
}
