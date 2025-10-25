package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"github.com/volchkovski/go-practicum-metrics/internal/logger"
	"io"
	"net/http"
)

func WithDecrypt(key *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Errorf("Failed to read request body: %s", err.Error())
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, key, data)
			if err != nil {
				logger.Log.Errorf("Failed to decrypt request body: %s", err.Error())
				http.Error(w, "Internal encryption", http.StatusBadRequest)
			}
			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			h.ServeHTTP(w, r)
		})
	}
}
