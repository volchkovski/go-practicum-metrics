package services

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
	ReadAllGauges() (map[string]float64, error)
	ReadAllCounters() (map[string]int64, error)
}

type MetricsReader interface {
	ReadGauge(string) (float64, error)
	ReadCounter(string) (int64, error)
}

type MetricsWriter interface {
	WriteGauge(string, float64) error
	WriteCounter(string, int64) error
}

type Pinger interface {
	Ping() error
}

type GaugesCountersWriter interface {
	WriteGaugesCounters(gauges map[string]float64, counters map[string]int64) error
}
