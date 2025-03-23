package collector

import (
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/storage"
	"net/http"
)

type Collector struct {
	mux *http.ServeMux
	s   storage.Storage
}

func (c Collector) Run() error {
	if err := http.ListenAndServe(`:8080`, c.mux); err != nil {
		return err
	}
	return nil
}

func New() Collector {
	mux := http.NewServeMux()
	s := storage.NewMemStorage()
	mux.Handle(`/update/`, handlers.CollectMetricHandler(s))
	return Collector{mux, s}
}
