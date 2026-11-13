package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		realIP     string
		want       string
	}{
		{"direct client", "203.0.113.7:1234", "", "203.0.113.7"},
		{"public peer cannot spoof X-Real-IP", "203.0.113.7:1234", "198.51.100.1", "203.0.113.7"},
		{"private proxy forwards X-Real-IP", "172.18.0.3:1234", "198.51.100.1", "198.51.100.1"},
		{"loopback proxy forwards X-Real-IP", "127.0.0.1:1234", "198.51.100.1", "198.51.100.1"},
		{"invalid X-Real-IP is ignored", "172.18.0.3:1234", "garbage", "172.18.0.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.realIP != "" {
				r.Header.Set("X-Real-IP", tt.realIP)
			}
			if got := ClientIP(r); got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(0.001, 2)
	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	do := func(remoteAddr string) int {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = remoteAddr
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}

	for i := 0; i < 2; i++ {
		if code := do("203.0.113.7:1"); code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200", i, code)
		}
	}
	if code := do("203.0.113.7:1"); code != http.StatusTooManyRequests {
		t.Fatalf("got %d, want 429", code)
	}
	if code := do("203.0.113.8:1"); code != http.StatusOK {
		t.Fatalf("other client: got %d, want 200", code)
	}
}
