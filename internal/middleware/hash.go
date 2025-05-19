package middleware

import (
	"bytes"
	"github.com/volchkovski/go-practicum-metrics/internal/hasher"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	"io"
	"net/http"
)

type cachedResponseWriter struct {
	http.ResponseWriter
	status int
	buff   *bytes.Buffer
}

func (crw *cachedResponseWriter) WriteHeader(status int) {
	crw.status = status
}

func (crw *cachedResponseWriter) Write(body []byte) (int, error) {
	return crw.buff.Write(body)
}

func WithHash(key string) func(http.Handler) http.Handler {
	hshr := hasher.New(key)
	return func(h http.Handler) http.Handler {
		hashFn := func(w http.ResponseWriter, r *http.Request) {
			reqHash := r.Header.Get(hasher.HashHeaderKey)
			if reqHash == "" {
				h.ServeHTTP(w, r)
				return
				//msg := fmt.Sprintf("Header %s is required", hasher.HashHeaderKey)
				//http.Error(w, msg, http.StatusBadRequest)
				//return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Errorf("Failed to read request body: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			valid, err := hshr.Validate(body, reqHash)
			if err != nil {
				logger.Log.Errorf("Failed to validate hash: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if !valid {
				http.Error(w, "Invalid hash", http.StatusBadRequest)
				return
			}
			buff := bytes.NewBuffer(nil)
			crw := &cachedResponseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				buff:           buff,
			}
			h.ServeHTTP(crw, r)
			respHash := hshr.Hash(buff.Bytes())
			w.Header().Set(hasher.HashHeaderKey, respHash)
			w.WriteHeader(crw.status)
			if _, err := io.Copy(w, buff); err != nil {
				logger.Log.Errorf("Failed to write response body: %v", err)
				http.Error(w, "Server internal error", http.StatusInternalServerError)
				return
			}
		}
		return http.HandlerFunc(hashFn)
	}
}
