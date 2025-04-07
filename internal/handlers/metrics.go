package handlers

import (
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

type MetricType string

const (
	GaugeType   = MetricType("gauge")
	CounterType = MetricType("counter")
)

type MetricGetter interface {
	GetGaugeMetric(string) (*m.GaugeMetric, error)
	GetCounterMetric(string) (*m.CounterMetric, error)
}

type MetricPusher interface {
	PushGaugeMetric(*m.GaugeMetric) error
	PushCounterMetric(*m.CounterMetric) error
}

type AllMetricsGetter interface {
	GetAllGaugeMetrics() ([]*m.GaugeMetric, error)
	GetAllCounterMetrics() ([]*m.CounterMetric, error)
}
