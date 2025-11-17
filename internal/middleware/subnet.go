package middleware

import (
	"net"
	"net/http"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

// WithTrustedSubnet creates a middleware that checks if the client's IP address
// is within the trusted subnet specified in CIDR notation.
// If trustedSubnet is empty, all requests are allowed.
// Otherwise, requests from IPs outside the trusted subnet receive a 403 Forbidden response.
func WithTrustedSubnet(trustedSubnet string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If no trusted subnet is configured, allow all requests
			if trustedSubnet == "" {
				h.ServeHTTP(w, r)
				return
			}

			// Parse the trusted subnet
			_, subnet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				logger.Log.Errorw("Invalid trusted subnet CIDR", "subnet", trustedSubnet, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Get the client IP from X-Real-IP header
			clientIP := r.Header.Get("X-Real-IP")
			if clientIP == "" {
				logger.Log.Warnw("X-Real-IP header is missing")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Parse the client IP
			ip := net.ParseIP(clientIP)
			if ip == nil {
				logger.Log.Warnw("Invalid IP address in X-Real-IP header", "ip", clientIP)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check if the IP is within the trusted subnet
			if !subnet.Contains(ip) {
				logger.Log.Warnw("IP address not in trusted subnet", "ip", clientIP, "subnet", trustedSubnet)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// IP is trusted, proceed with the request
			h.ServeHTTP(w, r)
		})
	}
}
