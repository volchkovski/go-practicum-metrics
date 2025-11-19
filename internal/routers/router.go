package routers

import (
	"crypto/rsa"
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

// NewMetricRouter creates a new HTTP router with metric endpoints.
// It configures routes for collecting and retrieving metrics with optional security middleware.
// The secretKey parameter enables hash-based authentication if provided.
// The trustedSubnet parameter restricts access to metric updates from specific IP ranges in CIDR notation.
// Debug endpoints (pprof) are conditionally added based on build tags.
func NewMetricRouter(secretKey string, rsaKey *rsa.PrivateKey, trustedSubnet string, s metricsProcessor) chi.Router {
	r := chi.NewRouter()
	r.Use(mw.WithLogging)
	if secretKey != "" {
		r.Use(mw.WithHash(secretKey))
	}
	r.With(mw.WithCompress).Get(`/`, handlers.AllMetricsHandler(s))
	r.Get(`/ping`, handlers.PingDB(s))
	r.Group(func(r chi.Router) {
		r.Use(mw.WithCompress)
		if trustedSubnet != "" {
			r.Use(mw.WithTrustedSubnet(trustedSubnet))
		}
		if rsaKey != nil {
			r.Use(mw.WithDecrypt(rsaKey))
		}
		r.Post(`/updates/`, handlers.CollectMetricsHandlerJSON(s))
	})
	r.Route(`/update`, func(r chi.Router) {
		if trustedSubnet != "" {
			r.Use(mw.WithTrustedSubnet(trustedSubnet))
		}
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

	// Add debug routes conditionally based on build tags
	addDebugRoutes(r)
	return r
}
