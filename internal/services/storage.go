package services

import "context"

type MetricStorage interface {
	MetricsReader
	MetricsWriter
	AllMetricsReader
	Pinger
	GaugesCountersWriter
	Closer
}

type Closer interface {
	Close() error
}

type AllMetricsReader interface {
	ReadAllGauges(context.Context) (map[string]float64, error)
	ReadAllCounters(context.Context) (map[string]int64, error)
}

type MetricsReader interface {
	ReadGauge(context.Context, string) (float64, error)
	ReadCounter(context.Context, string) (int64, error)
}

type MetricsWriter interface {
	WriteGauge(context.Context, string, float64) error
	WriteCounter(context.Context, string, int64) error
}

type Pinger interface {
	Ping(context.Context) error
}

type GaugesCountersWriter interface {
	WriteGaugesCounters(ctx context.Context, gauges map[string]float64, counters map[string]int64) error
}
