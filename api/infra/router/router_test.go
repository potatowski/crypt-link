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

func TestSetupWithoutRateLimit(t *testing.T) {
	h := SetupWithoutRateLimit(controller.NewMessageController(&stubService{}))

	// Far more requests than the in-memory limiter's burst: none may be rejected.
	for i := 0; i < 3*rateLimitBurst; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/message/"+testID, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, w.Code, http.StatusOK)
		}
		if got := w.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("Cache-Control = %q, want no-store", got)
		}
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/unknown", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown route: status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// TestRoutingEdgeCases pins the routing behaviour of the original Go 1.22
// patterns ("POST /api/message", "GET /api/message/{id}"), so the router keeps
// it on toolchains whose ServeMux lacks method and wildcard patterns (TinyGo).
func TestRoutingEdgeCases(t *testing.T) {
	svc := &stubService{}
	h := Setup(controller.NewMessageController(svc))

	tests := []struct {
		method, path string
		wantStatus   int
		wantID       string
	}{
		{http.MethodHead, "/api/message/" + testID, http.StatusMethodNotAllowed, ""},
		{http.MethodPut, "/api/message", http.StatusNotFound, ""},
		{http.MethodGet, "/api/message/", http.StatusNotFound, ""},
		{http.MethodGet, "/api/message/" + testID + "/extra", http.StatusNotFound, ""},
		{http.MethodGet, "/api/message/a%2Fb", http.StatusOK, "a/b"},
		{http.MethodGet, "/api/messages", http.StatusNotFound, ""},
		{http.MethodPost, "/api/message/" + testID, http.StatusNotFound, ""},
		{http.MethodGet, "/", http.StatusNotFound, ""},
		{http.MethodGet, "/api/message/" + testID, http.StatusOK, testID},
		{http.MethodGet, "/api/%6Dessage/abc", http.StatusOK, "abc"},
		{http.MethodGet, "/api/message/abc/", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			svc.gotID = ""
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if svc.gotID != tt.wantID {
				t.Fatalf("service got id %q, want %q", svc.gotID, tt.wantID)
			}
		})
	}
}

func TestCreateWithEscapedPath(t *testing.T) {
	h := Setup(controller.NewMessageController(&stubService{}))
	w := httptest.NewRecorder()
	body := fmt.Sprintf(`{"id":%q,"encrypted":"x"}`, testID)
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/%6Dessage", strings.NewReader(body)))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", w.Code, http.StatusCreated, w.Body.String())
	}
}

// Unclean paths are redirected to the cleaned path. Go 1.22+ answers 307;
// TinyGo's ServeMux (pre-1.22 behaviour) answers 301.
func TestRedirectsUncleanPaths(t *testing.T) {
	h := Setup(controller.NewMessageController(&stubService{}))
	for _, path := range []string{"/api/message/x/../abc", "/api//message/abc"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusTemporaryRedirect && w.Code != http.StatusMovedPermanently {
			t.Fatalf("%s: status = %d, want a redirect", path, w.Code)
		}
		if loc := w.Header().Get("Location"); loc != "/api/message/abc" {
			t.Fatalf("%s: Location = %q, want /api/message/abc", path, loc)
		}
	}
}
