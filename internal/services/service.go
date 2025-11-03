package services

import (
	"context"
	"fmt"

	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

// MetricService provides business logic for metric operations and validation.
// It acts as a layer between HTTP handlers and the storage backend, implementing
// proper error handling, context management, and business rules for metrics.
//
// The service supports two types of metrics:
//   - Gauge metrics: floating-point values that represent current state
//   - Counter metrics: integer values that can only increase
type MetricService struct {
	strg MetricStorage
}

// NewMetricService creates a new metric service with the provided storage backend.
// The storage parameter must implement MetricStorage interface for persistence operations.
func NewMetricService(strg MetricStorage) *MetricService {
	return &MetricService{strg}
}

// Close releases any resources held by the underlying storage.
// This method should be called when the service is no longer needed to prevent resource leaks.
func (ms *MetricService) Close() error {
	return ms.strg.Close()
}

// GetGaugeMetric retrieves a gauge metric by name from the storage.
// Returns an error if the metric is not found or if there's a storage issue.
// The context can be used to cancel the operation if it takes too long.
func (ms *MetricService) GetGaugeMetric(ctx context.Context, nm string) (*m.GaugeMetric, error) {
	val, err := ms.strg.ReadGauge(ctx, nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge metric with name %s: %w", nm, err)
	}
	return &m.GaugeMetric{Name: nm, Value: val}, nil
}

// GetCounterMetric retrieves a counter metric by name from the storage.
// Returns an error if the metric is not found or if there's a storage issue.
// The context can be used to cancel the operation if it takes too long.
func (ms *MetricService) GetCounterMetric(ctx context.Context, nm string) (*m.CounterMetric, error) {
	val, err := ms.strg.ReadCounter(ctx, nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter metric with name %s: %w", nm, err)
	}
	return &m.CounterMetric{Name: nm, Value: val}, nil
}

// PushGaugeMetric stores or updates a gauge metric in the storage.
// Gauge metrics can have their values completely replaced with new values.
// Returns an error if the storage operation fails.
func (ms *MetricService) PushGaugeMetric(ctx context.Context, m *m.GaugeMetric) error {
	if err := ms.strg.WriteGauge(ctx, m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push gauge metric with name name %s and value %.2f: %w", m.Name, m.Value, err)
	}
	return nil
}

// PushCounterMetric stores or updates a counter metric in the storage.
// Counter metrics are additive - the new value is added to the existing value.
// Returns an error if the storage operation fails.
func (ms *MetricService) PushCounterMetric(ctx context.Context, m *m.CounterMetric) error {
	if err := ms.strg.WriteCounter(ctx, m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push counter metric with name name %s and value %d: %w", m.Name, m.Value, err)
	}
	return nil
}

// GetAllGaugeMetrics retrieves all gauge metrics from the storage.
func (ms *MetricService) GetAllGaugeMetrics(ctx context.Context) ([]*m.GaugeMetric, error) {
	gauges, err := ms.strg.ReadAllGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauge metrics: %w", err)
	}
	gaugeMetrics := make([]*m.GaugeMetric, 0, 50)
	for nm, val := range gauges {
		gaugeMetrics = append(gaugeMetrics, &m.GaugeMetric{Name: nm, Value: val})
	}
	return gaugeMetrics, nil
}

// GetAllCounterMetrics retrieves all counter metrics from the storage.
func (ms *MetricService) GetAllCounterMetrics(ctx context.Context) ([]*m.CounterMetric, error) {
	counters, err := ms.strg.ReadAllCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter metrics: %w", err)
	}
	counterMetrics := make([]*m.CounterMetric, 0, 10)
	for nm, val := range counters {
		counterMetrics = append(counterMetrics, &m.CounterMetric{Name: nm, Value: val})
	}
	return counterMetrics, nil
}

// PingDB checks the health of the underlying storage connection.
func (ms *MetricService) PingDB(ctx context.Context) error {
	if err := ms.strg.Ping(ctx); err != nil {
		return fmt.Errorf("DB is not connected: %w", err)
	}
	return nil
}

// PushMetrics efficiently stores multiple gauge and counter metrics in a single operation.
func (ms *MetricService) PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
	gs := make(map[string]float64)
	cs := make(map[string]int64)

	for _, gauge := range gauges {
		gs[gauge.Name] = gauge.Value
	}
	for _, counter := range counters {
		cs[counter.Name] += counter.Value
	}

	if err := ms.strg.WriteGaugesCounters(ctx, gs, cs); err != nil {
		return fmt.Errorf("failed to write gauges and counters: %w", err)
	}
	return nil
}
