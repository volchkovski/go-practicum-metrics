package server

import (
	"net/http"

	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	router  http.Handler
	storage handlers.MetricsReadWriter
}

func (s *Server) Run() error {
	if err := http.ListenAndServe(`:8080`, s.router); err != nil {
		return err
	}
	return nil
}

func metricRouter(s handlers.MetricsReadWriter) chi.Router {
	r := chi.NewRouter()
	r.Get(`/`, handlers.AllMetricsHandler(s))
	r.Route(`/update/{tp}`, func(r chi.Router) {
		r.Post(`/`, http.NotFound)
		r.Post(`/{nm}/{val}`, handlers.CollectMetricHandler(s))
	})
	r.Get(`/value/{tp}/{nm}`, handlers.MetricHandler(s))
	return r
}

func New() *Server {
	s := storage.NewMemStorage()
	r := metricRouter(s)
	return &Server{r, s}
}
