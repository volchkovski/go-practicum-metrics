package agent

import (
	"github.com/go-resty/resty/v2"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

// HTTPClient представляет интерфейс для HTTP клиента
//
//go:generate go run go.uber.org/mock/mockgen -source=interfaces.go -destination=mocks_test.go -package=agent
type HTTPClient interface {
	R() *resty.Request
	SetTimeout(timeout interface{}) *resty.Client
}

// MetricsCollector представляет интерфейс для сбора метрик
type MetricsCollector interface {
	CollectRuntimeMetrics() map[string]float64
	CollectExtraMetrics() (map[string]float64, error)
}

// StorageInterface представляет интерфейс для хранения метрик
type StorageInterface interface {
	ReadMetrics() []*m.Metrics
	ReplaceMetrics(metrics ...*m.Metrics)
}
