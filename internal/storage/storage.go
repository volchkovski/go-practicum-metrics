package storage

import (
	"fmt"
	"sync"
)

type MemStorage struct {
	gauges       map[string]float64
	gaugesLock   sync.RWMutex
	counters     map[string]int64
	countersLock sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:       map[string]float64{},
		gaugesLock:   sync.RWMutex{},
		counters:     map[string]int64{},
		countersLock: sync.RWMutex{},
	}
}

func (s *MemStorage) WriteGauge(name string, value float64) error {
	s.gaugesLock.Lock()
	defer s.gaugesLock.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) WriteCounter(name string, value int64) error {
	s.countersLock.Lock()
	defer s.countersLock.Unlock()
	s.counters[name] += value
	return nil
}

func (s *MemStorage) ReadGauge(name string) (float64, error) {
	s.gaugesLock.RLock()
	defer s.gaugesLock.RUnlock()
	if m, ok := s.gauges[name]; ok {
		return m, nil
	}
	return 0, fmt.Errorf("%s not found", name)
}

func (s *MemStorage) ReadCounter(name string) (int64, error) {
	s.countersLock.RLock()
	defer s.countersLock.RUnlock()
	if m, ok := s.counters[name]; ok {
		return m, nil
	}
	return 0, fmt.Errorf("%s not found", name)
}

func (s *MemStorage) ReadAllGauges() (map[string]float64, error) {
	s.gaugesLock.RLock()
	defer s.gaugesLock.RUnlock()
	gauges := s.gauges
	return gauges, nil
}

func (s *MemStorage) ReadAllCounters() (map[string]int64, error) {
	s.countersLock.RLock()
	defer s.countersLock.RUnlock()
	counters := s.counters
	return counters, nil
}
