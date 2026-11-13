package controller

import (
	"context"
	"crypt-link/core/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testID = "3f2b8c1e-9d4a-4b6f-8e2a-1c5d7f9a0b3e"

type fakeService struct {
	createErr error
	getMsg    *domain.Message
	getErr    error

	gotID string
}

func (f *fakeService) Create(_ context.Context, id, _ string, _ time.Time) error {
	f.gotID = id
	return f.createErr
}

func (f *fakeService) GetAndInvalidate(_ context.Context, id string) (*domain.Message, error) {
	f.gotID = id
	return f.getMsg, f.getErr
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	return body.Error
}

func TestCreateMessage(t *testing.T) {
	validBody := fmt.Sprintf(`{"id":%q,"encrypted":"secret"}`, testID)

	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantError  string
	}{
		{"created", validBody, nil, http.StatusCreated, ""},
		{"invalid JSON", `{"id":`, nil, http.StatusBadRequest, "invalid request body"},
		{"unknown field", fmt.Sprintf(`{"id":%q,"encrypted":"x","extra":1}`, testID), nil, http.StatusBadRequest, "invalid request body"},
		{"trailing data", validBody + `{}`, nil, http.StatusBadRequest, "invalid request body"},
		{"body too large", fmt.Sprintf(`{"id":%q,"encrypted":"%s"}`, testID, strings.Repeat("a", MaxRequestBodyBytes)), nil, http.StatusRequestEntityTooLarge, "request body too large"},
		{"invalid input", validBody, fmt.Errorf("%w: bad id", domain.ErrInvalidInput), http.StatusBadRequest, "invalid message data"},
		{"internal error is not leaked", validBody, errors.New("mongo: connection refused"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := NewMessageController(&fakeService{createErr: tt.serviceErr})
			r := httptest.NewRequest(http.MethodPost, "/api/message", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			ctrl.CreateMessage(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if tt.wantError != "" {
				if got := decodeError(t, w); got != tt.wantError {
					t.Errorf("error = %q, want %q", got, tt.wantError)
				}
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name       string
		msg        *domain.Message
		serviceErr error
		wantStatus int
		wantError  string
	}{
		{"found", &domain.Message{ID: testID, Encrypted: "cipher"}, nil, http.StatusOK, ""},
		{"not found", nil, domain.ErrNotFound, http.StatusNotFound, "message not found or expired"},
		{"expired", nil, domain.ErrExpired, http.StatusNotFound, "message not found or expired"},
		{"invalid id", nil, fmt.Errorf("%w: bad id", domain.ErrInvalidInput), http.StatusBadRequest, "invalid message data"},
		{"internal error is not leaked", nil, errors.New("mongo: timeout"), http.StatusInternalServerError, "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{getMsg: tt.msg, getErr: tt.serviceErr}
			ctrl := NewMessageController(svc)
			r := httptest.NewRequest(http.MethodGet, "/api/message/"+testID, nil)
			r.SetPathValue("id", testID)
			w := httptest.NewRecorder()

			ctrl.GetMessage(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if svc.gotID != testID {
				t.Errorf("service got id %q, want %q", svc.gotID, testID)
			}
			if tt.wantError != "" {
				if got := decodeError(t, w); got != tt.wantError {
					t.Errorf("error = %q, want %q", got, tt.wantError)
				}
				return
			}
			var body map[string]string
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if body["encrypted"] != "cipher" {
				t.Errorf("encrypted = %q, want cipher", body["encrypted"])
			}
		})
	}
}

func TestGetMessageMissingID(t *testing.T) {
	ctrl := NewMessageController(&fakeService{})
	w := httptest.NewRecorder()
	ctrl.GetMessage(w, httptest.NewRequest(http.MethodGet, "/api/message/", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
