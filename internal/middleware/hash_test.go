package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/volchkovski/go-practicum-metrics/internal/hasher"
)

func TestCachedResponseWriter(t *testing.T) {
	t.Run("write header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		buff := bytes.NewBuffer(nil)
		crw := &cachedResponseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
			buff:           buff,
		}

		crw.WriteHeader(http.StatusCreated)
		assert.Equal(t, http.StatusCreated, crw.status)
		// Underlying ResponseWriter header should not be called yet
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("write body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		buff := bytes.NewBuffer(nil)
		crw := &cachedResponseWriter{
			ResponseWriter: rec,
			status:         http.StatusOK,
			buff:           buff,
		}

		testData := []byte("test response body")
		n, err := crw.Write(testData)

		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, buff.Bytes())
		// Data should not be in underlying ResponseWriter yet
		assert.Empty(t, rec.Body.Bytes())
	})
}

func TestWithHash(t *testing.T) {
	key := "test-secret-key"
	hshr := hasher.New(key)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "error reading body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response: " + string(body)))
	})

	hashHandler := WithHash(key)(handler)

	t.Run("request without hash header", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", strings.NewReader("test body"))
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		// Should pass through without hash validation
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: test body", rec.Body.String())
	})

	t.Run("request with valid hash", func(t *testing.T) {
		requestBody := "test request body"
		hash := hshr.Hash([]byte(requestBody))

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, hash)
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: "+requestBody, rec.Body.String())

		// Response should have hash header
		responseHash := rec.Header().Get(hasher.HashHeaderKey)
		assert.NotEmpty(t, responseHash)

		// Validate response hash
		expectedHash := hshr.Hash([]byte("response: " + requestBody))
		assert.Equal(t, expectedHash, responseHash)
	})

	t.Run("request with invalid hash", func(t *testing.T) {
		requestBody := "test request body"
		invalidHash := "invalid-hash"

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, invalidHash)
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid hash")
	})

	t.Run("request with wrong hash for body", func(t *testing.T) {
		requestBody := "test request body"
		wrongBody := "different body"
		hash := hshr.Hash([]byte(wrongBody))

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, hash)
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid hash")
	})

	t.Run("empty request body with valid hash", func(t *testing.T) {
		requestBody := ""
		hash := hshr.Hash([]byte(requestBody))

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, hash)
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: ", rec.Body.String())

		// Response should have hash header
		responseHash := rec.Header().Get(hasher.HashHeaderKey)
		assert.NotEmpty(t, responseHash)
	})

	t.Run("handler that writes custom status code", func(t *testing.T) {
		customHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte("created: " + string(body)))
		})

		customHashHandler := WithHash(key)(customHandler)

		requestBody := "test body"
		hash := hshr.Hash([]byte(requestBody))

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, hash)
		rec := httptest.NewRecorder()

		customHashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "created: "+requestBody, rec.Body.String())

		// Response should have hash header
		responseHash := rec.Header().Get(hasher.HashHeaderKey)
		assert.NotEmpty(t, responseHash)

		// Validate response hash
		expectedHash := hshr.Hash([]byte("created: " + requestBody))
		assert.Equal(t, expectedHash, responseHash)
	})

	t.Run("large request body", func(t *testing.T) {
		requestBody := strings.Repeat("a", 10000) // 10KB body
		hash := hshr.Hash([]byte(requestBody))

		req := httptest.NewRequest("POST", "/", strings.NewReader(requestBody))
		req.Header.Set(hasher.HashHeaderKey, hash)
		rec := httptest.NewRecorder()

		hashHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "response: "+requestBody, rec.Body.String())

		// Response should have hash header
		responseHash := rec.Header().Get(hasher.HashHeaderKey)
		assert.NotEmpty(t, responseHash)
	})
}
