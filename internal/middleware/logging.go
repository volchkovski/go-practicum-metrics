package middleware

import (
	"net/http"
	"time"

	"github.com/volchkovski/go-practicum-metrics/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *loggingResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		logger.Log.Infow(
			"Got request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
		)
		logger.Log.Infow(
			"Sent response",
			"status", lw.status,
			"size", lw.size,
		)
	}
	return http.HandlerFunc(logFn)
}
