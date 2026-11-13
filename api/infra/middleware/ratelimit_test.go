package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

func TestVisitorKey(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{"203.0.113.7", "203.0.113.7"},
		{"::ffff:203.0.113.7", "203.0.113.7"},
		{"2001:db8:1:2:aaaa:bbbb:cccc:dddd", "2001:db8:1:2::/64"},
		{"garbage", "garbage"},
	}
	for _, tt := range tests {
		if got := visitorKey(tt.ip); got != tt.want {
			t.Errorf("visitorKey(%q) = %q, want %q", tt.ip, got, tt.want)
		}
	}
}

func TestRateLimiterIPv6SamePrefixSharesBucket(t *testing.T) {
	rl := NewRateLimiter(0.001, 1)
	if !rl.Allow("2001:db8:1:2::1") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("2001:db8:1:2:ffff:ffff:ffff:ffff") {
		t.Fatal("address in the same /64 should share the exhausted bucket")
	}
}

func TestRateLimiterIPv6DifferentPrefixSeparateBuckets(t *testing.T) {
	rl := NewRateLimiter(0.001, 1)
	if !rl.Allow("2001:db8:1:2::1") {
		t.Fatal("first request should be allowed")
	}
	if !rl.Allow("2001:db8:1:3::1") {
		t.Fatal("address in a different /64 should have its own bucket")
	}
}

func TestRateLimiterIPv4MappedSharesBucketWithIPv4(t *testing.T) {
	rl := NewRateLimiter(0.001, 1)
	if !rl.Allow("203.0.113.7") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("::ffff:203.0.113.7") {
		t.Fatal("IPv4-mapped IPv6 should share the IPv4 bucket")
	}
}

func TestRateLimiterVisitorCap(t *testing.T) {
	rl := NewRateLimiter(0.001, 1)
	rl.maxVisitors = 3
	now := time.Now()
	rl.now = func() time.Time { return now }

	for i := 1; i <= 10; i++ {
		ip := fmt.Sprintf("203.0.113.%d", i)
		if !rl.Allow(ip) {
			t.Fatalf("first request from %s should be allowed", ip)
		}
	}
	if got := len(rl.visitors); got != 3 {
		t.Fatalf("visitors = %d, want 3", got)
	}
	// Untracked clients fail open; tracked ones are still limited.
	if !rl.Allow("203.0.113.10") {
		t.Fatal("untracked client should be allowed when the map is full")
	}
	if rl.Allow("203.0.113.1") {
		t.Fatal("tracked client should still be rate limited")
	}

	// After the idle TTL, the next sweep frees room for new clients.
	now = now.Add(visitorIdleTTL + cleanupInterval + time.Second)
	rl.Allow("203.0.113.20")
	if got := len(rl.visitors); got != 1 {
		t.Fatalf("visitors after sweep = %d, want 1", got)
	}
}
