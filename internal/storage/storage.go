package storage

import (
	"fmt"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (s *MemStorage) WriteGauge(name string, value float64) error {
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) WriteCounter(name string, value int64) error {
	s.counters[name] += value
	return nil
}

func (s *MemStorage) ReadGauge(name string) (float64, error) {
	if m, ok := s.gauges[name]; ok {
		return m, nil
	}
	return 0, fmt.Errorf("%s not found", name)
}

func (s *MemStorage) ReadCounter(name string) (int64, error) {
	if m, ok := s.counters[name]; ok {
		return m, nil
	}
	return 0, fmt.Errorf("%s not found", name)
}

func (s *MemStorage) ReadAllGauges() map[string]float64 {
	return s.gauges
}

func (s *MemStorage) ReadAllCounters() map[string]int64 {
	return s.counters
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   map[string]float64{},
		counters: map[string]int64{},
	}
}
