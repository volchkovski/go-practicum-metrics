package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	router chi.Router
	addr   string
	notify chan error
}

func New(r chi.Router, addr string) *HTTPServer {
	return &HTTPServer{
		router: r,
		addr:   addr,
		notify: make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		s.notify <- http.ListenAndServe(s.addr, s.router)
		close(s.notify)
	}()
}

func (s *HTTPServer) Notify() chan error {
	return s.notify
}
