package public

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/orbito/api-go/internal/shared/httpx"
)

type rateBucket struct {
	count   int
	resetAt time.Time
}

// RateLimiter enforces per-IP fixed-window rate limits matching NestJS PublicRateLimitGuard.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
	nowFn   func() time.Time
}

// NewRateLimiter initializes an IP-keyed rate limiter for public endpoints.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*rateBucket),
		nowFn:   time.Now,
	}
}

// LimitForCategory resolves endpoint-category rate limits (requests per window).
func LimitForCategory(category string) (limit int, window time.Duration) {
	window = time.Minute
	switch category {
	case "search":
		return 10, window
	case "shared_attachment":
		return 5, window
	case "shared_task":
		return 10, window
	case "projects":
		return 50, window
	case "tasks":
		return 30, window
	case "sprints":
		return 30, window
	case "calendar":
		return 20, window
	default:
		return 100, window
	}
}

// Middleware applies rate limiting for a given endpoint bucket category (or custom limit/window when limit > 0).
func (rl *RateLimiter) Middleware(category string, customLimit int, customWindow time.Duration) func(http.Handler) http.Handler {
	limit, window := LimitForCategory(category)
	if customLimit > 0 {
		limit = customLimit
	}
	if customWindow > 0 {
		window = customWindow
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := extractClientIP(r)
			key := category + ":" + ip

			allowed, remaining, resetAt, retryAfter := rl.take(key, limit, window)
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", resetAt.UTC().Format(time.RFC3339))

			if !allowed {
				httpx.RespondJSON(w, http.StatusTooManyRequests, map[string]any{
					"message":    "Rate limit exceeded for public endpoint. Please try again later.",
					"error":      "Too Many Requests",
					"statusCode": http.StatusTooManyRequests,
					"retryAfter": retryAfter,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) take(key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time, retryAfter int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.nowFn()
	if len(rl.buckets) > 10000 {
		rl.pruneLocked(now)
	}

	b, ok := rl.buckets[key]
	if !ok || !now.Before(b.resetAt) {
		resetAt = now.Add(window)
		rl.buckets[key] = &rateBucket{count: 1, resetAt: resetAt}
		return true, max(0, limit-1), resetAt, 0
	}

	if b.count >= limit {
		secs := int(b.resetAt.Sub(now).Seconds())
		if secs < 1 {
			secs = 1
		}
		return false, 0, b.resetAt, secs
	}

	b.count++
	return true, max(0, limit-b.count), b.resetAt, 0
}

func (rl *RateLimiter) pruneLocked(now time.Time) {
	for k, v := range rl.buckets {
		if !now.Before(v.resetAt) {
			delete(rl.buckets, k)
		}
	}
}

func extractClientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
