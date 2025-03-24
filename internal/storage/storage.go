package storage

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Storage interface {
	WriteGauge(string, float64) error
	WriteCounter(string, int64) error
}

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

func (s *MemStorage) WriteGauge(name string, value float64) error {
	s.GS.Gauges[name] = value
	return nil
}

func (s *MemStorage) WriteCounter(name string, value int64) error {
	s.CS.Counters[name] += value
	return nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		GS: &GaugeStorage{
			Gauges: map[string]float64{},
		},
		CS: &CounterStorage{
			Counters: map[string]int64{},
		},
	}
}
