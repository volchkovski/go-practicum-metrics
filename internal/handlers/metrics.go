package handlers

import (
	"context"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

type MetricType string

type Metric struct {
	Name  string
	Value string
}

const (
	GaugeType   = MetricType("gauge")
	CounterType = MetricType("counter")
)

type MetricGetter interface {
	GetGaugeMetric(context.Context, string) (*m.GaugeMetric, error)
	GetCounterMetric(context.Context, string) (*m.CounterMetric, error)
}

type MetricPusher interface {
	PushGaugeMetric(context.Context, *m.GaugeMetric) error
	PushCounterMetric(context.Context, *m.CounterMetric) error
}

type MetricsPusher interface {
	PushMetrics(context.Context, []*m.GaugeMetric, []*m.CounterMetric) error
}

type AllMetricsGetter interface {
	GetAllGaugeMetrics(context.Context) ([]*m.GaugeMetric, error)
	GetAllCounterMetrics(context.Context) ([]*m.CounterMetric, error)
}

type DBPinger interface {
	PingDB(context.Context) error
}
