package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	"github.com/volchkovski/go-practicum-metrics/internal/middleware"
)

type metricsProcessor interface {
	handlers.MetricGetter
	handlers.MetricPusher
	handlers.AllMetricsGetter
}

func NewMetricRouter(s metricsProcessor) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.WithLogging)
	r.Get(`/`, handlers.AllMetricsHandler(s))
	r.Route(`/update`, func(r chi.Router) {
		r.Post(`/`, handlers.CollectMetricHandlerJSON(s))
		r.Route(`/{tp}`, func(r chi.Router) {
			r.Post(`/`, http.NotFound)
			r.Post(`/{nm}/{val}`, handlers.CollectMetricHandler(s))
		})
	})
	r.Route(`/value`, func(r chi.Router) {
		r.Post(`/`, handlers.MetricHandlerJSON(s))
		r.Get(`/{tp}/{nm}`, handlers.MetricHandler(s))
	})
	return r
}
