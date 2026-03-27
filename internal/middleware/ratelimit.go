package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/HMasataka/gate/internal/domain"
)

// RateLimit returns a middleware that applies rate limiting per IP address per endpoint.
// limit: max requests per window. Key is scoped to clientIP + request path.
func RateLimit(limiter domain.RateLimiter, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getRealIP(r)
			key := fmt.Sprintf("ip:%s:%s", clientIP, r.URL.Path)

			allowed, err := limiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				// On limiter error, fail open to avoid blocking legitimate traffic.
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				writeRateLimitExceeded(w, window)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeRateLimitExceeded(w http.ResponseWriter, window time.Duration) {
	retryAfter := int(window.Seconds())
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":    "rate_limit_exceeded",
			"message": "too many requests",
		},
	})
}

// getRealIP extracts the real client IP from the request.
// It checks X-Forwarded-For first (proxy-aware), then falls back to RemoteAddr.
func getRealIP(r *http.Request) string {
	// X-Forwarded-For may contain a comma-separated list; the first entry is the client IP.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	// X-Real-IP is set by some proxies (e.g. nginx).
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}

	// Strip port from RemoteAddr.
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}

	return addr
}
