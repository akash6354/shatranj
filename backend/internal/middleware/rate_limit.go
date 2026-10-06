package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/shatranj/backend/internal/httpapi"
)

type rateEntry struct {
	start time.Time
	count int
}

const maxTrackedRateLimitClients = 10000

// RateLimit limits requests per remote IP over a fixed window.
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	if limit < 1 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	var mu sync.Mutex
	entries := make(map[string]rateEntry)
	lastCleanup := time.Now()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			now := time.Now()
			mu.Lock()
			entry, exists := entries[ip]
			if !exists && len(entries) >= maxTrackedRateLimitClients {
				for address, old := range entries {
					if now.Sub(old.start) >= window {
						delete(entries, address)
					}
				}
				if len(entries) >= maxTrackedRateLimitClients {
					mu.Unlock()
					w.Header().Set("Retry-After", strconv.Itoa(max(1, int(window.Seconds()))))
					httpapi.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
					return
				}
			}
			if entry.start.IsZero() || now.Sub(entry.start) >= window {
				entry = rateEntry{start: now}
			}
			entry.count++
			entries[ip] = entry
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

			if !allowed {
				retryAfter := int((window - now.Sub(entry.start)).Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				httpapi.WriteError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
