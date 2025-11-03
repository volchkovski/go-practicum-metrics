// Package models contains data structures for metrics collection and storage.
package models

// GaugeMetric represents a gauge-type metric with a floating-point value.
type GaugeMetric struct {
	Name  string  `json:"name"`  // Name is the unique identifier for the metric
	Value float64 `json:"value"` // Value is the current floating-point value of the metric
}

// CounterMetric represents a counter-type metric with an integer value.
type CounterMetric struct {
	Name  string `json:"name"`  // Name is the unique identifier for the metric
	Value int64  `json:"value"` // Value is the current integer value of the counter
}
