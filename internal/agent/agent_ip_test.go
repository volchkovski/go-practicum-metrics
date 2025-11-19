package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	// This test verifies that getLocalIP returns a valid IP address
	// or empty string if no suitable IP is found
	ip := getLocalIP()

	// The IP should either be empty or a valid IPv4 address
	if ip != "" {
		// Check that it's not a loopback address
		assert.NotEqual(t, "127.0.0.1", ip, "IP should not be loopback")

		// Basic validation that it looks like an IP address
		// (contains dots for IPv4)
		assert.Contains(t, ip, ".", "IP should contain dots (IPv4)")

		t.Logf("Found local IP: %s", ip)
	} else {
		t.Log("No local IP found (this is acceptable in some environments)")
	}
}
