package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTrustedSubnet(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "Empty trusted subnet allows all",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP address",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "invalid-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "IPv6 in trusted subnet",
			trustedSubnet:  "2001:db8::/32",
			xRealIP:        "2001:db8::1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IPv6 not in trusted subnet",
			trustedSubnet:  "2001:db8::/32",
			xRealIP:        "2001:db9::1",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := WithTrustedSubnet(tt.trustedSubnet)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "/updates/", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
