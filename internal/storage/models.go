package storage

type MemStorage struct {
	GS *GaugeStorage
	CS *CounterStorage
}

type GaugeStorage struct {
	Gauges map[string]float64
}

type CounterStorage struct {
	Counters map[string]int64
}
