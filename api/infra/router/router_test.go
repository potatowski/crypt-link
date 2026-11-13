package router

import (
	"context"
	"crypt-link/adapter/input/controller"
	"crypt-link/core/domain"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testID = "3f2b8c1e-9d4a-4b6f-8e2a-1c5d7f9a0b3e"

type stubService struct{ gotID string }

func (s *stubService) Create(context.Context, string, string, int64) error { return nil }

func (s *stubService) GetAndInvalidate(_ context.Context, id string) (*domain.Message, error) {
	s.gotID = id
	return &domain.Message{ID: id, Encrypted: "cipher"}, nil
}

func TestRoutes(t *testing.T) {
	svc := &stubService{}
	h := Setup(controller.NewMessageController(svc))

	tests := []struct {
		method, path, body string
		wantStatus         int
	}{
		{http.MethodPost, "/api/message", fmt.Sprintf(`{"id":%q,"encrypted":"x"}`, testID), http.StatusCreated},
		{http.MethodGet, "/api/message/" + testID, "", http.StatusOK},
		{http.MethodGet, "/api/unknown", "", http.StatusNotFound},
		{http.MethodDelete, "/api/message/" + testID, "", http.StatusNotFound},
		{http.MethodGet, "/api/message", "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if got := w.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
			}
		})
	}

	if svc.gotID != testID {
		t.Errorf("path value id = %q, want %q", svc.gotID, testID)
	}
}

func TestRateLimitApplied(t *testing.T) {
	h := Setup(controller.NewMessageController(&stubService{}))

	var last int
	for i := 0; i < rateLimitBurst+1; i++ {
		r := httptest.NewRequest(http.MethodGet, "/api/message/"+testID, nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		last = w.Code
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("status after burst = %d, want 429", last)
	}
}
