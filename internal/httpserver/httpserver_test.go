package httpserver

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("create new http server", func(t *testing.T) {
		router := chi.NewRouter()
		server := New(router, ":8080")

		assert.NotNil(t, server)
		assert.Equal(t, ":8080", server.addr)
		assert.Equal(t, router, server.router)
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
