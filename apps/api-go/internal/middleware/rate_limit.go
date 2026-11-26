package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// RateLimiter implements a thread-safe fixed/sliding-window per-key rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	entries map[string]*rateWindow
}

type rateWindow struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter creates a rate limiter permitting max requests per window.
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:     max,
		window:  window,
		entries: make(map[string]*rateWindow),
	}
}

// Allow checks and increments the request counter for the provided key.
func (rl *RateLimiter) Allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if len(rl.entries) > 10000 {
		for k, v := range rl.entries {
			if now.After(v.resetAt) {
				delete(rl.entries, k)
			}
		}
	}

	w, exists := rl.entries[key]
	if !exists || now.After(w.resetAt) {
		rl.entries[key] = &rateWindow{count: 1, resetAt: now.Add(rl.window)}
		return true
	}
	if w.count >= rl.max {
		return false
	}
	w.count++
	return true
}

// UserRateLimit enforces the limiter keyed by authenticated user ID (or remote IP fallback).
func UserRateLimit(max int, window time.Duration) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(max, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ctxhelper.GetUserID(r.Context())
			if key == "" {
				key = r.RemoteAddr
			}
			if !limiter.Allow(key, time.Now()) {
				httpx.RespondError(w, r, apperror.New(
					http.StatusTooManyRequests,
					"Too Many Requests",
					"Rate limit exceeded. Please try again later.",
				))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
