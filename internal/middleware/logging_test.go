package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingResponseWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	lw := &loggingResponseWriter{
		ResponseWriter: recorder,
		status:         0,
		size:           0,
	}

	testData := []byte("test response data")
	n, err := lw.Write(testData)

	require.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, len(testData), lw.size)
	assert.Equal(t, string(testData), recorder.Body.String())
}

func TestLoggingResponseWriter_Write_Multiple(t *testing.T) {
	recorder := httptest.NewRecorder()
	lw := &loggingResponseWriter{
		ResponseWriter: recorder,
		status:         0,
		size:           0,
	}

	writes := [][]byte{
		[]byte("first "),
		[]byte("second "),
		[]byte("third"),
	}

	totalSize := 0
	for _, data := range writes {
		n, err := lw.Write(data)
		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		totalSize += len(data)
	}

	assert.Equal(t, totalSize, lw.size)
	assert.Equal(t, "first second third", recorder.Body.String())
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	tests := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusInternalServerError,
		http.StatusNotFound,
	}

	for _, statusCode := range tests {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			lw := &loggingResponseWriter{
				ResponseWriter: recorder,
				status:         0,
				size:           0,
			}

			lw.WriteHeader(statusCode)
			assert.Equal(t, statusCode, lw.status)
			assert.Equal(t, statusCode, recorder.Code)
		})
	}
}

func TestWithLogging(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	loggedHandler := WithLogging(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test/path?param=value", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test response", recorder.Body.String())
}

func TestWithLogging_DifferentMethods(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("method: " + r.Method))
	})

	loggedHandler := WithLogging(testHandler)

	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			recorder := httptest.NewRecorder()

			loggedHandler.ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusOK, recorder.Code)
			assert.Equal(t, "method: "+method, recorder.Body.String())
		})
	}
}

func TestWithLogging_DifferentStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, status := range statusCodes {
		t.Run(http.StatusText(status), func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				w.Write([]byte("response body"))
			})

			loggedHandler := WithLogging(testHandler)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			recorder := httptest.NewRecorder()

			loggedHandler.ServeHTTP(recorder, req)

			assert.Equal(t, status, recorder.Code)
			assert.Equal(t, "response body", recorder.Body.String())
		})
	}
}

func TestWithLogging_EmptyResponse(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	loggedHandler := WithLogging(testHandler)
	req := httptest.NewRequest(http.MethodPost, "/empty", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Body.String())
}

func TestWithLogging_LargeResponse(t *testing.T) {
	largeData := strings.Repeat("large response data ", 1000)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeData))
	})

	loggedHandler := WithLogging(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/large", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, largeData, recorder.Body.String())
}

func TestWithLogging_MultipleWrites(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("part1"))
		w.Write([]byte("part2"))
		w.Write([]byte("part3"))
	})

	loggedHandler := WithLogging(testHandler)
	req := httptest.NewRequest(http.MethodGet, "/multi", nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "part1part2part3", recorder.Body.String())
}

func TestWithLogging_ComplexURL(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	loggedHandler := WithLogging(testHandler)

	complexURL := "/api/v1/metrics/gauge/cpu_usage/75.5?timestamp=123456&source=agent"
	req := httptest.NewRequest(http.MethodPost, complexURL, nil)
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "success", recorder.Body.String())
}

func TestWithLogging_WithBody(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})

	loggedHandler := WithLogging(testHandler)

	body := strings.NewReader(`{"id":"test","type":"gauge","value":123.45}`)
	req := httptest.NewRequest(http.MethodPost, "/update", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	loggedHandler.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code)
	assert.Equal(t, "created", recorder.Body.String())
}
