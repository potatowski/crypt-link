package middleware

import (
	"crypt-link/response"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// visitorIdleTTL is how long an idle client's limiter is kept in memory.
	visitorIdleTTL = 3 * time.Minute
	// cleanupInterval is the minimum time between sweeps of idle limiters.
	cleanupInterval = time.Minute
)

var errTooManyRequests = errors.New("too many requests")

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is an in-memory, per-client-IP token bucket rate limiter.
type RateLimiter struct {
	mu          sync.Mutex
	visitors    map[string]*visitor
	limit       rate.Limit
	burst       int
	lastCleanup time.Time
	now         func() time.Time
}

// NewRateLimiter creates a limiter allowing rps requests per second per client
// IP, with bursts of up to burst requests.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
		now:      time.Now,
	}
}

// Allow reports whether a request from the given client IP may proceed.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	if now.Sub(rl.lastCleanup) > cleanupInterval {
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > visitorIdleTTL {
				delete(rl.visitors, k)
			}
		}
		rl.lastCleanup = now
	}

	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = now
	return v.limiter.AllowN(now, 1)
}

// Middleware rejects requests exceeding the rate limit with 429 Too Many Requests.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(ClientIP(r)) {
			w.Header().Set("Retry-After", "1")
			response.Error(w, http.StatusTooManyRequests, errTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP returns the client IP for the request. The X-Real-IP header (set
// by the nginx reverse proxy) is honored only when the direct peer is a
// loopback or private address, so it cannot be spoofed by clients reaching
// the API directly. X-Forwarded-For is intentionally ignored.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	peer := net.ParseIP(host)
	if peer != nil && (peer.IsLoopback() || peer.IsPrivate()) {
		if real := net.ParseIP(r.Header.Get("X-Real-IP")); real != nil {
			return real.String()
		}
	}
	return host
}
