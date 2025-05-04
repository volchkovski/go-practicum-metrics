package services

import (
	"fmt"

	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

type MetricService struct {
	repo MetricsReadWriter
	db   Pinger
}

func NewMetricService(repo MetricsReadWriter, db Pinger) *MetricService {
	return &MetricService{
		repo: repo,
		db:   db,
	}
}

func (ms *MetricService) GetGaugeMetric(nm string) (*m.GaugeMetric, error) {
	val, err := ms.repo.ReadGauge(nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge metric with name %s: %w", nm, err)
	}
	return &m.GaugeMetric{Name: nm, Value: val}, nil
}

func (ms *MetricService) GetCounterMetric(nm string) (*m.CounterMetric, error) {
	val, err := ms.repo.ReadCounter(nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter metric with name %s: %w", nm, err)
	}
	return &m.CounterMetric{Name: nm, Value: val}, nil
}

func (ms *MetricService) PushGaugeMetric(m *m.GaugeMetric) error {
	if err := ms.repo.WriteGauge(m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push gauge metric with name name %s and value %.2f: %w", m.Name, m.Value, err)
	}
	return nil
}

func (ms *MetricService) PushCounterMetric(m *m.CounterMetric) error {
	if err := ms.repo.WriteCounter(m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push counter metric with name name %s and value %d: %w", m.Name, m.Value, err)
	}
	return nil
}

func (ms *MetricService) GetAllGaugeMetrics() ([]*m.GaugeMetric, error) {
	gauges, err := ms.repo.ReadAllGauges()
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauge metrics: %w", err)
	}
	gaugeMetrics := make([]*m.GaugeMetric, 0, 50)
	for nm, val := range gauges {
		gaugeMetrics = append(gaugeMetrics, &m.GaugeMetric{Name: nm, Value: val})
	}
	return gaugeMetrics, nil
}

func (ms *MetricService) GetAllCounterMetrics() ([]*m.CounterMetric, error) {
	counters, err := ms.repo.ReadAllCounters()
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter metrics: %w", err)
	}
	counterMetrics := make([]*m.CounterMetric, 0, 10)
	for nm, val := range counters {
		counterMetrics = append(counterMetrics, &m.CounterMetric{Name: nm, Value: val})
	}
	return counterMetrics, nil
}

func (ms *MetricService) PingDB() error {
	if ms.db == nil {
		return fmt.Errorf("DB is not initialized")
	}
	if err := ms.db.Ping(); err != nil {
		return fmt.Errorf("DB is not connected: %w", err)
	}
	return nil
}
