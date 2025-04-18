package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
)

type metricsProcessor interface {
	handlers.MetricGetter
	handlers.MetricPusher
	handlers.AllMetricsGetter
}

func NewMetricRouter(s metricsProcessor) chi.Router {
	r := chi.NewRouter()
	r.Get(`/`, handlers.AllMetricsHandler(s))
	r.Route(`/update/{tp}`, func(r chi.Router) {
		r.Post(`/`, http.NotFound)
		r.Post(`/{nm}/{val}`, handlers.CollectMetricHandler(s))
	})
	r.Get(`/value/{tp}/{nm}`, handlers.MetricHandler(s))
	return r
}
