package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/volchkovski/go-practicum-metrics/internal/handlers"
	mw "github.com/volchkovski/go-practicum-metrics/internal/middleware"
)

type metricsProcessor interface {
	handlers.MetricGetter
	handlers.MetricPusher
	handlers.AllMetricsGetter
	handlers.MetricsPusher
	handlers.DBPinger
}

func NewMetricRouter(s metricsProcessor) chi.Router {
	r := chi.NewRouter()
	r.Use(mw.WithLogging)
	r.With(mw.WithCompress).Get(`/`, handlers.AllMetricsHandler(s))
	r.Get(`/ping`, handlers.PingDB(s))
	r.With(mw.WithCompress).Post(`/updates/`, handlers.CollectMetricsHandlerJSON(s))
	r.Route(`/update`, func(r chi.Router) {
		r.With(mw.WithCompress).Post(`/`, handlers.CollectMetricHandlerJSON(s))
		r.Route(`/{tp}`, func(r chi.Router) {
			r.Post(`/`, http.NotFound)
			r.Post(`/{nm}/{val}`, handlers.CollectMetricHandler(s))
		})
	})
	r.Route(`/value`, func(r chi.Router) {
		r.With(mw.WithCompress).Post(`/`, handlers.MetricHandlerJSON(s))
		r.Get(`/{tp}/{nm}`, handlers.MetricHandler(s))
	})
	return r
}
