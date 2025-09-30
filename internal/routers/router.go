package routers

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

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

// NewMetricRouter creates a new HTTP router with metric endpoints.
// It configures routes for collecting and retrieving metrics with optional security middleware.
// The secretKey parameter enables hash-based authentication if provided.
func NewMetricRouter(secretKey string, s metricsProcessor) chi.Router {
	r := chi.NewRouter()
	r.Use(mw.WithLogging)
	if secretKey != "" {
		r.Use(mw.WithHash(secretKey))
	}
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

	r.Mount("/debug", middleware.Profiler())
	return r
}
