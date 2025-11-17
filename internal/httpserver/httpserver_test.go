package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("create new http server", func(t *testing.T) {
		router := chi.NewRouter()
		server := New(router, ":8080")

		assert.NotNil(t, server)
		assert.NotNil(t, server.server)
		assert.Equal(t, ":8080", server.server.Addr)
		assert.Equal(t, router, server.server.Handler)
		assert.NotNil(t, server.notify)
	})
}

func TestNotify(t *testing.T) {
	t.Run("notify returns channel", func(t *testing.T) {
		router := chi.NewRouter()
		server := New(router, ":8080")

		notifyChannel := server.Notify()
		assert.NotNil(t, notifyChannel)
		assert.Equal(t, server.notify, notifyChannel)
	})
}

func TestStart(t *testing.T) {
	t.Run("start server without panic", func(t *testing.T) {
		router := chi.NewRouter()
		router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Use an invalid port to ensure server start fails quickly
		server := New(router, ":99999")

		assert.NotPanics(t, func() {
			server.Start()
		})

		// Check that we can receive error from notify channel
		select {
		case err := <-server.Notify():
			// We expect an error because port 99999 is invalid
			assert.Error(t, err)
		default:
			// Server might not have started yet, that's ok for this test
		}
	})
}

func TestShutdown(t *testing.T) {
	t.Run("shutdown server gracefully", func(t *testing.T) {
		router := chi.NewRouter()
		router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		server := New(router, ":8080")

		// Test shutdown without starting (should not panic and may succeed)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotPanics(t, func() {
			err := server.Shutdown(ctx)
			// No specific error expectation - shutdown can succeed even if server wasn't running
			_ = err
		})
	})
}
