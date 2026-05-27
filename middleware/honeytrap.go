package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"sananti/core"
)

// HoneyTrap coordinates Core Blocker and Logger services to deploy
// deceptive honeypots across application routes.
type HoneyTrap struct {
	blocker core.Blocker
	logger  core.Logger
}

// NewHoneyTrap initializes a HoneyTrap component with the required core capabilities.
func NewHoneyTrap(blocker core.Blocker, logger core.Logger) *HoneyTrap {
	return &HoneyTrap{
		blocker: blocker,
		logger:  logger,
	}
}

// HandleTrap returns a middleware that intercepts requests targeting a specific targetURL.
// Intercepted requests are logged, blacklisted, and closed with a randomized 403 or 404 response.
func (ht *HoneyTrap) HandleTrap(targetURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == targetURL {
				ip := ExtractIP(r)
				reason := fmt.Sprintf("Triggered decoy honeytoken URL: %s", targetURL)

				// Create log alert metadata
				alert := core.AlertData{
					IP:        ip,
					Path:      r.URL.Path,
					Method:    r.Method,
					UserAgent: r.UserAgent(),
					Timestamp: time.Now(),
					Details:   reason,
				}
				
				// Push the alert to our asynchronous file logger channel
				_ = ht.logger.LogAlert(alert)

				// Block the IP in our thread-safe blacklist engine
				_ = ht.blocker.BlockIP(ip, reason)

				// Randomize between 403 Forbidden and 404 Not Found to obfuscate defense configurations
				var b [1]byte
				_, _ = rand.Read(b[:])
				if b[0]%2 == 0 {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.Header().Set("X-Content-Type-Options", "nosniff")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte("403 Forbidden\n"))
				} else {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.Header().Set("X-Content-Type-Options", "nosniff")
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("404 page not found\n"))
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ProtectionMiddleware returns a global middleware that checks every request against the blacklist.
// Blacklisted IPs are immediately intercepted and served a 403 Forbidden response.
func (ht *HoneyTrap) ProtectionMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ExtractIP(r)
			blocked, _, err := ht.blocker.IsBlocked(ip)
			if err == nil && blocked {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("403 Forbidden\n"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
