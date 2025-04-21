package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HttpServer struct {
	router chi.Router
	addr   string
	notify chan error
}

func New(r chi.Router, addr string) *HttpServer {
	return &HttpServer{
		router: r,
		addr:   addr,
		notify: make(chan error, 1),
	}
}

func (s *HttpServer) Start() {
	go func() {
		s.notify <- http.ListenAndServe(s.addr, s.router)
		close(s.notify)
	}()
}

func (s *HttpServer) Notify() chan error {
	return s.notify
}
