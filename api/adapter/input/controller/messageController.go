package controller

import (
	"crypt-link/core/domain"
	"crypt-link/core/port"
	"crypt-link/response"
	"encoding/json"
	"errors"
	"log/slog"
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

	// MaxRequestBodyBytes limits the size of request bodies.
	MaxRequestBodyBytes = 64 * 1024
)

var (
	errInvalidBody = errors.New("invalid request body")
	errBodyTooBig  = errors.New("request body too large")
	errInvalidData = errors.New("invalid message data")
	errNotFound    = errors.New("message not found or expired")
	errInternal    = errors.New("internal server error")
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
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createRequest
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			response.Error(w, http.StatusRequestEntityTooLarge, errBodyTooBig)
			return
		}
		response.Error(w, http.StatusBadRequest, errInvalidBody)
		return
	}
	if dec.More() {
		response.Error(w, http.StatusBadRequest, errInvalidBody)
		return
	}

	expiresAt := time.Now().Add(ValidityTime)

	if err := c.service.Create(r.Context(), req.ID, req.Encrypted, expiresAt); err != nil {
		writeServiceError(w, err)
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

	msg, err := c.service.GetAndInvalidate(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"encrypted": msg.Encrypted})
}

// writeServiceError maps domain errors to HTTP status codes without leaking
// internal error details to the client.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		response.Error(w, http.StatusBadRequest, errInvalidData)
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, domain.ErrExpired):
		response.Error(w, http.StatusNotFound, errNotFound)
	default:
		slog.Error("message service error", "error", err)
		response.Error(w, http.StatusInternalServerError, errInternal)
	}
}
