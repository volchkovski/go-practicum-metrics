package handlers

type MetricType string

const (
	GaugeType   = MetricType("gauge")
	CounterType = MetricType("counter")
)

type MetricsReadWriter interface {
	MetricsReader
	MetricsWriter
	AllMetricsReader
}

type AllMetricsReader interface {
	ReadAllGauges() map[string]float64
	ReadAllCounters() map[string]int64
}

type MetricsWriter interface {
	GaugeWriter
	CounterWriter
}

type MetricsReader interface {
	GaugeReader
	CounterReader
}

type GaugeWriter interface {
	WriteGauge(string, float64) error
}

type CounterWriter interface {
	WriteCounter(string, int64) error
}

type GaugeReader interface {
	ReadGauge(string) (float64, error)
}

type CounterReader interface {
	ReadCounter(string) (int64, error)
}
