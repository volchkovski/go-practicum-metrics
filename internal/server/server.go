package server

import (
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
	"net/http"
)

type Server struct {
	mux *http.ServeMux
	s   storage.Storage
}

func (s *Server) Run() error {
	if err := http.ListenAndServe(`:8080`, s.mux); err != nil {
		return err
	}
	return nil
}

func New() *Server {
	mux := http.NewServeMux()
	s := storage.NewMemStorage()
	mux.HandleFunc(`/update/`, handlers.CollectMetricHandler(s))
	return &Server{mux, s}
}
