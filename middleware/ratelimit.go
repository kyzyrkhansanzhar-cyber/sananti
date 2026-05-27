package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"sananti/core"
)

// IPBucket holds the token bucket state for a single IP address.
type IPBucket struct {
	tokens     float64
	lastRefill time.Time
}

// RateLimiter manages thread-safe token bucket rate limiting for all incoming client IPs.
type RateLimiter struct {
	mu         sync.Mutex
	ips        map[string]*IPBucket
	rate       float64 // Tokens refilled per second
	capacity   float64 // Maximum tokens held by the bucket (burst)
	blocker    core.Blocker
	logger     core.Logger
}

// NewRateLimiter initializes a rate limiter with the specified rate (req/sec) and capacity (burst size).
func NewRateLimiter(rate float64, capacity float64, blocker core.Blocker, logger core.Logger) *RateLimiter {
	return &RateLimiter{
		ips:      make(map[string]*IPBucket),
		rate:     rate,
		capacity: capacity,
		blocker:  blocker,
		logger:   logger,
	}
}

// LimitMiddleware intercepts incoming requests and returns a 429 status code if the rate limit is exceeded.
func (rl *RateLimiter) LimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractIP(r)

			rl.mu.Lock()
			bucket, exists := rl.ips[ip]
			now := time.Now()

			if !exists {
				bucket = &IPBucket{
					tokens:     rl.capacity,
					lastRefill: now,
				}
				rl.ips[ip] = bucket
			} else {
				// Refill tokens based on time elapsed since last refill
				elapsed := now.Sub(bucket.lastRefill).Seconds()
				bucket.tokens += elapsed * rl.rate
				if bucket.tokens > rl.capacity {
					bucket.tokens = rl.capacity
				}
				bucket.lastRefill = now
			}

			// Check if we have at least one token to consume
			if bucket.tokens >= 1.0 {
				bucket.tokens -= 1.0
				rl.mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}
			rl.mu.Unlock()

			// Rate limit exceeded: log an alert, block the IP temporarily, and return 429
			reason := fmt.Sprintf("Rate limit exceeded (burst capacity limit of %.0f reached)", rl.capacity)
			alert := core.AlertData{
				IP:        ip,
				Path:      r.URL.Path,
				Method:    r.Method,
				UserAgent: r.UserAgent(),
				Severity:  core.SeverityWarning,
				Timestamp: now,
				Details:   reason,
			}
			_ = rl.logger.LogAlert(alert)
			_ = rl.blocker.BlockIP(ip, reason)

			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("429 Too Many Requests: Rate limit exceeded\n"))
		})
	}
}
