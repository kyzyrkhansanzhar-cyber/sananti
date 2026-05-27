package middleware

import (
	"net"
	"net/http"
	"strings"
)

// ExtractIP extracts the real IP address of the client from HTTP headers,
// accounting for proxies, load balancers, and direct connections.
func ExtractIP(r *http.Request) string {
	// 1. Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain a list of comma-separated IPs.
		// The first one is typically the real client.
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// 2. Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if isValidIP(ip) {
			return ip
		}
	}

	// 3. Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If RemoteAddr doesn't have a port (e.g. in tests)
		ip = r.RemoteAddr
	}
	return ip
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
