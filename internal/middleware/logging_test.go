package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingResponseWriter(t *testing.T) {
	t.Run("write method", func(t *testing.T) {
		rec := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: rec,
			status:         0,
			size:           0,
		}

		data := []byte("test response data")
		n, err := lrw.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, len(data), lrw.size)
		assert.Equal(t, data, rec.Body.Bytes())
	})

	t.Run("write header", func(t *testing.T) {
		rec := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: rec,
			status:         0,
			size:           0,
		}

		lrw.WriteHeader(http.StatusNotFound)

		assert.Equal(t, http.StatusNotFound, lrw.status)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("default status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		lrw := &loggingResponseWriter{
			ResponseWriter: rec,
			status:         0,
			size:           0,
		}

		// Write without calling WriteHeader should default to 200
		data := []byte("test")
		lrw.Write(data)

		// Status should remain 0 until WriteHeader is called
		assert.Equal(t, 0, lrw.status)
		assert.Equal(t, len(data), lrw.size)
	})
}

func TestWithLogging(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("test response"))
	})

	loggingHandler := WithLogging(handler)

	t.Run("logs request and response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		loggingHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "test response", rec.Body.String())
	})

	t.Run("logs POST request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/metrics", nil)
		rec := httptest.NewRecorder()

		loggingHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "test response", rec.Body.String())
	})

	t.Run("logs request with query parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?param=value", nil)
		rec := httptest.NewRecorder()

		loggingHandler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}
