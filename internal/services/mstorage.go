package services

type MetricsReadWriter interface {
	MetricsReader
	MetricsWriter
	AllMetricsReader
}

type AllMetricsReader interface {
	ReadAllGauges() (map[string]float64, error)
	ReadAllCounters() (map[string]int64, error)
}

type MetricsWriter interface {
	WriteGauge(string, float64) error
	WriteCounter(string, int64) error
}

type MetricsReader interface {
	ReadGauge(string) (float64, error)
	ReadCounter(string) (int64, error)
}
