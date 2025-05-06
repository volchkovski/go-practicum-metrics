package services

type PgWriter interface {
	GaugesCountersWriter
	Pinger
}

type Pinger interface {
	Ping() error
}

type GaugesCountersWriter interface {
	WriteGaugesCounters(gauges map[string]float64, counters map[string]int64) error
}
