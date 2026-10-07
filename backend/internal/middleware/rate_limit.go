package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/shatranj/backend/internal/cache"
	"github.com/shatranj/backend/internal/httpapi"
)

type rateEntry struct {
	start time.Time
	count int
}

const maxTrackedRateLimitClients = 10000

// RateLimit limits requests per remote IP over a fixed window.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	return RateLimitWithStore(nil, "default", limit, window)
}

// RateLimitWithStore limits requests per remote IP using Redis/cache when
// available and falls back to process memory if the cache operation fails.
func RateLimitWithStore(store cache.Store, scope string, limit int, window time.Duration) func(http.Handler) http.Handler {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	if scope == "" {
		scope = "default"
	}
	var mu sync.Mutex
	entries := make(map[string]rateEntry)
	lastCleanup := time.Now()
	fallback := func(identity string) (bool, int) {
		now := time.Now()
		mu.Lock()
		entry, exists := entries[identity]
		if !exists && len(entries) >= maxTrackedRateLimitClients {
			for address, old := range entries {
				if now.Sub(old.start) >= window {
					delete(entries, address)
				}
			}
			if len(entries) >= maxTrackedRateLimitClients {
				mu.Unlock()
				return false, max(1, int(window.Seconds()))
			}
		}
		if entry.start.IsZero() || now.Sub(entry.start) >= window {
			entry = rateEntry{start: now}
		}
		entry.count++
		entries[identity] = entry
		allowed := entry.count <= limit
		if now.Sub(lastCleanup) > window {
			for address, old := range entries {
				if now.Sub(old.start) >= window {
					delete(entries, address)
				}
			}
			lastCleanup = now
		}
		mu.Unlock()
		retryAfter := int((window - now.Sub(entry.start)).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return allowed, retryAfter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			allowed, retryAfter := false, 1
			if limiter, ok := store.(cache.RateLimiterStore); ok {
				count, remaining, err := limiter.Increment(r.Context(), cache.RateLimitKey(scope, ip), window)
				if err == nil {
					allowed = count <= int64(limit)
					retryAfter = int(remaining.Seconds())
					if retryAfter < 1 {
						retryAfter = 1
					}
				} else {
					allowed, retryAfter = fallback(ip)
				}
			} else {
				allowed, retryAfter = fallback(ip)
			}
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				httpapi.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
