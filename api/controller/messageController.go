package controller

import (
	"crypt-link/service"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type createRequest struct {
	ID        string `json:"id"`
	Encrypted string `json:"encrypted"`
}

func CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var inner map[string]interface{}
	if err := json.Unmarshal([]byte(req.Encrypted), &inner); err != nil {
		http.Error(w, "Invalid encrypted content", http.StatusBadRequest)
		return
	}

	expiresStr, ok := inner["expiresAt"].(string)
	if !ok {
		http.Error(w, "expiresAt required", http.StatusBadRequest)
		return
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresStr)
	if err != nil {
		http.Error(w, "Invalid expiresAt", http.StatusBadRequest)
		return
	}

	service := service.NewMessageService()
	if err := service.Create(req.ID, req.Encrypted, expiresAt); err != nil {
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	service := service.NewMessageService()
	msg, err := service.GetAndInvalidate(id)
	if err != nil {
		http.Error(w, "Message not found or expired", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"encrypted": msg.Encrypted})
}
