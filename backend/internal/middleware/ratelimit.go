package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements per-IP rate limiting.
type RateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rateLimiterEntry
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter.
// rate: requests per second
// burst: maximum burst size
func NewRateLimiter(r float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*rateLimiterEntry),
		rate:     rate.Limit(r),
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// getLimiter returns the rate limiter for the given IP.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = &rateLimiterEntry{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	entry.lastSeen = time.Now()
	return entry.limiter
}

// cleanupLoop removes old entries periodically.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		for ip, entry := range rl.limiters {
			if time.Since(entry.lastSeen) > rl.cleanup {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Limit middleware applies rate limiting to a handler.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for CORS preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		ip := getClientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request.
// Properly handles X-Forwarded-For to prevent IP spoofing.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from reverse proxy)
	// Take the first IP in the chain (original client IP)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			// First IP is the original client
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// Fall back to RemoteAddr, stripping port if present
	remoteAddr := r.RemoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		// Check if this is IPv6 (contains brackets)
		if !strings.Contains(remoteAddr, "]") || strings.LastIndex(remoteAddr, ":") > strings.LastIndex(remoteAddr, "]") {
			remoteAddr = remoteAddr[:idx]
		}
	}
	return remoteAddr
}
