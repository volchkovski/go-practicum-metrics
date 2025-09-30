package handlers

import (
	"context"

	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

// MetricType represents the type of metric (gauge or counter).
type MetricType string

// Metric represents a metric with its name and string value for display purposes.
type Metric struct {
	Name  string // Name is the metric identifier
	Value string // Value is the string representation of the metric value
}

const (
	// GaugeType represents gauge metrics that can increase or decrease.
	GaugeType = MetricType("gauge")
	// CounterType represents counter metrics that can by only increase.
	CounterType = MetricType("counter")
)

// MetricGetter defines the interface for retrieving individual metrics.
type MetricGetter interface {
	// GetGaugeMetric retrieves a gauge metric by name.
	GetGaugeMetric(context.Context, string) (*m.GaugeMetric, error)
	// GetCounterMetric retrieves a counter metric by name.
	GetCounterMetric(context.Context, string) (*m.CounterMetric, error)
}

// MetricPusher defines the interface for storing individual metrics.
type MetricPusher interface {
	// PushGaugeMetric stores a gauge metric.
	PushGaugeMetric(context.Context, *m.GaugeMetric) error
	// PushCounterMetric stores a counter metric.
	PushCounterMetric(context.Context, *m.CounterMetric) error
}

// MetricsPusher defines the interface for storing multiple metrics in a batch.
type MetricsPusher interface {
	// PushMetrics stores multiple gauge and counter metrics in a single operation.
	PushMetrics(context.Context, []*m.GaugeMetric, []*m.CounterMetric) error
}

// AllMetricsGetter defines the interface for retrieving all stored metrics.
type AllMetricsGetter interface {
	// GetAllGaugeMetrics retrieves all stored gauge metrics.
	GetAllGaugeMetrics(context.Context) ([]*m.GaugeMetric, error)
	// GetAllCounterMetrics retrieves all stored counter metrics.
	GetAllCounterMetrics(context.Context) ([]*m.CounterMetric, error)
}

// DBPinger defines the interface for checking database connectivity.
type DBPinger interface {
	// PingDB checks if the database connection is healthy.
	PingDB(context.Context) error
}
