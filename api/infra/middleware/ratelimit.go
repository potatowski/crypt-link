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
	// maxVisitors caps the number of tracked clients to bound memory.
	maxVisitors = 100_000
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
	maxVisitors int
	now         func() time.Time
}

// NewRateLimiter creates a limiter allowing rps requests per second per client
// IP, with bursts of up to burst requests.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		visitors:    make(map[string]*visitor),
		limit:       rate.Limit(rps),
		burst:       burst,
		maxVisitors: maxVisitors,
		now:         time.Now,
	}
}

// Allow reports whether a request from the given client IP may proceed.
// Clients are bucketed by visitorKey (IPv6 per /64, IPv4 per address).
func (rl *RateLimiter) Allow(ip string) bool {
	key := visitorKey(ip)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	// The O(n) idle sweep runs at most once per cleanupInterval, never on
	// every request.
	if now.Sub(rl.lastCleanup) > cleanupInterval {
		for k, v := range rl.visitors {
			if now.Sub(v.lastSeen) > visitorIdleTTL {
				delete(rl.visitors, k)
			}
		}
		rl.lastCleanup = now
	}

	v, ok := rl.visitors[key]
	if !ok {
		// When the map is full (and the periodic sweep has not freed room),
		// fail open: allow the request without tracking it. This keeps memory
		// bounded and costs O(1) per request; the trade-off is that new
		// clients are unlimited until the next sweep evicts idle entries.
		if len(rl.visitors) >= rl.maxVisitors {
			return true
		}
		v = &visitor{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.visitors[key] = v
	}
	v.lastSeen = now
	return v.limiter.AllowN(now, 1)
}

// visitorKey maps a client IP to its rate-limit bucket key. IPv4 addresses
// (including IPv4-mapped IPv6) are keyed per address; other IPv6 addresses are
// keyed by their /64 prefix, since a single client typically controls a whole
// /64 and could otherwise rotate addresses to bypass the limit.
func visitorKey(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	return parsed.Mask(net.CIDRMask(64, 128)).String() + "/64"
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
