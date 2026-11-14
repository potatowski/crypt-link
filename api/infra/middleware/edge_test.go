package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEdgeHeaders(t *testing.T) {
	h := EdgeHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/message", nil))

	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusTeapot)
	}
	want := map[string]string{
		"Content-Security-Policy":   "default-src 'none'; frame-ancestors 'none'",
		"Referrer-Policy":           "no-referrer",
		"Strict-Transport-Security": "max-age=31536000",
	}
	for k, v := range want {
		if got := w.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
}
