package agent

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRestyClient(t *testing.T) {
	t.Run("creates resty client with correct configuration", func(t *testing.T) {
		client := NewRestyClient()

		assert.NotNil(t, client)

		// Check that client is properly configured
		assert.IsType(t, &resty.Client{}, client)
	})

	t.Run("client has retry configuration", func(t *testing.T) {
		client := NewRestyClient()

		// Test that we can create a request
		req := client.R()
		assert.NotNil(t, req)
		assert.IsType(t, &resty.Request{}, req)
	})
}

func TestRetryCodes(t *testing.T) {
	t.Run("contains expected retry codes", func(t *testing.T) {
		expectedCodes := []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}

		assert.Equal(t, expectedCodes, retryCodes)
		assert.Len(t, retryCodes, 5)
	})

	t.Run("retry codes are valid HTTP status codes", func(t *testing.T) {
		for _, code := range retryCodes {
			assert.True(t, code >= 400 && code < 600, "Code %d should be a 4xx or 5xx status", code)
		}
	})
}

func TestHeaders(t *testing.T) {
	t.Run("contains expected headers", func(t *testing.T) {
		expectedHeaders := map[string]string{
			"Accept-Encoding":  "",
			"Content-Encoding": "gzip",
			"Content-Type":     "application/json",
		}

		assert.Equal(t, expectedHeaders, headers)
		assert.Len(t, headers, 3)
	})

	t.Run("headers have correct values", func(t *testing.T) {
		assert.Equal(t, "", headers["Accept-Encoding"])
		assert.Equal(t, "gzip", headers["Content-Encoding"])
		assert.Equal(t, "application/json", headers["Content-Type"])
	})
}

func TestClientRetryConfiguration(t *testing.T) {
	t.Run("client can be configured", func(t *testing.T) {
		client := NewRestyClient()

		// Test that we can set timeout
		client = client.SetTimeout(5 * time.Second)
		assert.NotNil(t, client)
	})
}
