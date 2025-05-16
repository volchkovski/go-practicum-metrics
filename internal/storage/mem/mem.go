package mem

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
)

var ErrCanceled = errors.New("operation is canceled")

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

func (s *MemStorage) Close() error {
	return nil
}

func (s *MemStorage) WriteGauge(ctx context.Context, name string, value float64) error {
	select {
	case <-ctx.Done():
		return ErrCanceled
	default:
		s.gaugesLock.Lock()
		defer s.gaugesLock.Unlock()
		s.gauges[name] = value
		return nil
	}
}

func (s *MemStorage) WriteCounter(ctx context.Context, name string, value int64) error {
	select {
	case <-ctx.Done():
		return ErrCanceled
	default:
		s.countersLock.Lock()
		defer s.countersLock.Unlock()
		s.counters[name] += value
		return nil
	}
}

func (s *MemStorage) ReadGauge(ctx context.Context, name string) (float64, error) {
	select {
	case <-ctx.Done():
		return 0, ErrCanceled
	default:
		s.gaugesLock.RLock()
		defer s.gaugesLock.RUnlock()
		if m, ok := s.gauges[name]; ok {
			return m, nil
		}
		return 0, fmt.Errorf("%s not found", name)
	}
}

func (s *MemStorage) ReadCounter(ctx context.Context, name string) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ErrCanceled
	default:
		s.countersLock.RLock()
		defer s.countersLock.RUnlock()
		if m, ok := s.counters[name]; ok {
			return m, nil
		}
		return 0, fmt.Errorf("%s not found", name)
	}
}

func (s *MemStorage) ReadAllGauges(ctx context.Context) (map[string]float64, error) {
	select {
	case <-ctx.Done():
		return nil, ErrCanceled
	default:
		s.gaugesLock.RLock()
		defer s.gaugesLock.RUnlock()
		gauges := s.gauges
		return gauges, nil
	}
}

func (s *MemStorage) ReadAllCounters(ctx context.Context) (map[string]int64, error) {
	select {
	case <-ctx.Done():
		return nil, ErrCanceled
	default:
		s.countersLock.RLock()
		defer s.countersLock.RUnlock()
		counters := s.counters
		return counters, nil
	}
}

func (s *MemStorage) WriteGaugesCounters(ctx context.Context, gauges map[string]float64, counters map[string]int64) error {
	select {
	case <-ctx.Done():
		return ErrCanceled
	default:
		s.gaugesLock.Lock()
		defer s.gaugesLock.Unlock()
		maps.Copy(s.gauges, gauges)

		s.countersLock.Lock()
		defer s.countersLock.Unlock()
		for nm, v := range counters {
			s.counters[nm] += v
		}
		return nil
	}
}

func (s *MemStorage) Ping(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrCanceled
	default:
		return nil
	}
}
