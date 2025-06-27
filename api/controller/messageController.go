package controller

import (
	"crypt-link/response"
	"crypt-link/service"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type createRequest struct {
	ID           string `json:"id"`
	Encrypted    string `json:"encrypted"`
	ValidityTime int    `json:"validityTime"`
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err)
		return
	}

	if req.ValidityTime <= 1 || req.ValidityTime > 1440 {
		response.Error(w, http.StatusBadRequest, errors.New("validityTime must be between 1 and 1440 minutes"))
		return
	}

	expiresAt := time.Now().Add(time.Duration(req.ValidityTime) * time.Minute)

	service := service.NewMessageService()
	if err := service.Create(req.ID, req.Encrypted, expiresAt); err != nil {
		response.Error(w, http.StatusInternalServerError, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "Message created successfully"})
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		response.Error(w, http.StatusBadRequest, errors.New("ID is required"))
		return
	}

	service := service.NewMessageService()
	msg, err := service.GetAndInvalidate(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, errors.New("message not found or expired"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"encrypted": msg.Encrypted})
}
