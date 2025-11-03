package services

import "context"

// MetricStorage defines the complete interface for metric storage operations.
// It combines all the required interfaces for a full-featured metric storage system.
type MetricStorage interface {
	MetricsReader
	MetricsWriter
	AllMetricsReader
	Pinger
	GaugesCountersWriter
	Closer
}

// Closer defines the interface for resources that need cleanup.
type Closer interface {
	// Close releases any resources held by the storage implementation.
	Close() error
}

// AllMetricsReader defines the interface for reading all metrics of each type.
type AllMetricsReader interface {
	// ReadAllGauges returns all gauge metrics as a map of name to value.
	ReadAllGauges(context.Context) (map[string]float64, error)
	// ReadAllCounters returns all counter metrics as a map of name to value.
	ReadAllCounters(context.Context) (map[string]int64, error)
}

// MetricsReader defines the interface for reading individual metrics.
type MetricsReader interface {
	// ReadGauge retrieves the value of a gauge metric by name.
	ReadGauge(context.Context, string) (float64, error)
	// ReadCounter retrieves the value of a counter metric by name.
	ReadCounter(context.Context, string) (int64, error)
}

// MetricsWriter defines the interface for writing individual metrics.
type MetricsWriter interface {
	// WriteGauge stores a gauge metric with the given name and value.
	WriteGauge(context.Context, string, float64) error
	// WriteCounter stores a counter metric with the given name and value.
	WriteCounter(context.Context, string, int64) error
}

// Pinger defines the interface for checking storage connectivity.
type Pinger interface {
	// Ping verifies that the storage backend is accessible and responsive.
	Ping(context.Context) error
}

// GaugesCountersWriter defines the interface for batch writing metrics.
type GaugesCountersWriter interface {
	// WriteGaugesCounters efficiently stores multiple gauges and counters in a single operation.
	WriteGaugesCounters(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
}
