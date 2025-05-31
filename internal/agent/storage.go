package agent

import (
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"slices"
	"sync"
)

type MetricsStorage struct {
	metrics []*m.Metrics
	mu      sync.RWMutex
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		metrics: make([]*m.Metrics, 0, 50),
	}
}

func (ms *MetricsStorage) ReadMetrics() []*m.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return slices.Clone(ms.metrics)
}

func (ms *MetricsStorage) ReplaceMetrics(metrics ...*m.Metrics) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.metrics = metrics
}
