// Package httpserver provides HTTP server functionality with graceful shutdown support.
package httpserver

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	server *http.Server
	notify chan error
}

func New(r chi.Router, addr string) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
		notify: make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *HTTPServer) Notify() chan error {
	return s.notify
}
