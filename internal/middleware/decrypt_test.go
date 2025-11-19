package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

func init() {
	// Initialize logger for tests
	_ = logger.Initialize("info", "local")
}

func TestWithDecrypt(t *testing.T) {
	// Generate RSA key pair for tests
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	t.Run("successfully decrypts valid encrypted data", func(t *testing.T) {
		plaintext := []byte("test data")
		
		// Encrypt data
		ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, plaintext)
		require.NoError(t, err)

		// Create handler that checks decrypted body
		var receivedBody []byte
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		})

		// Apply decrypt middleware
		middleware := WithDecrypt(privateKey)
		wrappedHandler := middleware(handler)

		// Create test request
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(ciphertext))
		rec := httptest.NewRecorder()

		// Execute
		wrappedHandler.ServeHTTP(rec, req)

		// Verify
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, plaintext, receivedBody)
	})

	t.Run("handles read error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called")
		})

		middleware := WithDecrypt(privateKey)
		wrappedHandler := middleware(handler)

		// Create request with error reader
		req := httptest.NewRequest(http.MethodPost, "/test", errorReader{})
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "Internal server error")
	})

	t.Run("handles decryption error with invalid data", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called on decryption error")
		})

		middleware := WithDecrypt(privateKey)
		wrappedHandler := middleware(handler)

		// Send unencrypted/invalid data
		invalidData := []byte("not encrypted data")
		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(invalidData))
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		// The middleware returns 400 BadRequest on decryption error
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("handles empty body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("handler should not be called on decryption error")
		})

		middleware := WithDecrypt(privateKey)
		wrappedHandler := middleware(handler)

		req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte{}))
		rec := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(rec, req)

		// Empty body causes decryption error and returns 400
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("can decrypt multiple requests", func(t *testing.T) {
		plaintext1 := []byte("first message")
		plaintext2 := []byte("second message")
		
		ciphertext1, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, plaintext1)
		require.NoError(t, err)
		ciphertext2, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, plaintext2)
		require.NoError(t, err)

		receivedBodies := make([][]byte, 0, 2)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			receivedBodies = append(receivedBodies, body)
			w.WriteHeader(http.StatusOK)
		})

		middleware := WithDecrypt(privateKey)
		wrappedHandler := middleware(handler)

		// First request
		req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(ciphertext1))
		rec1 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec1, req1)

		// Second request
		req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(ciphertext2))
		rec2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec2, req2)

		assert.Equal(t, 2, len(receivedBodies))
		assert.Equal(t, plaintext1, receivedBodies[0])
		assert.Equal(t, plaintext2, receivedBodies[1])
	})
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

