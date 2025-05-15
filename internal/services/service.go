package services

import (
	"context"
	"fmt"

	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

type MetricService struct {
	strg MetricStorage
}

func (ms *MetricService) Close() error {
	return ms.strg.Close()
}

func NewMetricService(strg MetricStorage) *MetricService {
	return &MetricService{strg}
}

func (ms *MetricService) GetGaugeMetric(ctx context.Context, nm string) (*m.GaugeMetric, error) {
	val, err := ms.strg.ReadGauge(ctx, nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get gauge metric with name %s: %w", nm, err)
	}
	return &m.GaugeMetric{Name: nm, Value: val}, nil
}

func (ms *MetricService) GetCounterMetric(ctx context.Context, nm string) (*m.CounterMetric, error) {
	val, err := ms.strg.ReadCounter(ctx, nm)
	if err != nil {
		return nil, fmt.Errorf("failed to get counter metric with name %s: %w", nm, err)
	}
	return &m.CounterMetric{Name: nm, Value: val}, nil
}

func (ms *MetricService) PushGaugeMetric(ctx context.Context, m *m.GaugeMetric) error {
	if err := ms.strg.WriteGauge(ctx, m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push gauge metric with name name %s and value %.2f: %w", m.Name, m.Value, err)
	}
	return nil
}

func (ms *MetricService) PushCounterMetric(ctx context.Context, m *m.CounterMetric) error {
	if err := ms.strg.WriteCounter(ctx, m.Name, m.Value); err != nil {
		return fmt.Errorf("failed to push counter metric with name name %s and value %d: %w", m.Name, m.Value, err)
	}
	return nil
}

func (ms *MetricService) GetAllGaugeMetrics(ctx context.Context) ([]*m.GaugeMetric, error) {
	gauges, err := ms.strg.ReadAllGauges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauge metrics: %w", err)
	}
	gaugeMetrics := make([]*m.GaugeMetric, 0, 50)
	for nm, val := range gauges {
		gaugeMetrics = append(gaugeMetrics, &m.GaugeMetric{Name: nm, Value: val})
	}
	return gaugeMetrics, nil
}

func (ms *MetricService) GetAllCounterMetrics(ctx context.Context) ([]*m.CounterMetric, error) {
	counters, err := ms.strg.ReadAllCounters(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter metrics: %w", err)
	}
	counterMetrics := make([]*m.CounterMetric, 0, 10)
	for nm, val := range counters {
		counterMetrics = append(counterMetrics, &m.CounterMetric{Name: nm, Value: val})
	}
	return counterMetrics, nil
}

func (ms *MetricService) PingDB(ctx context.Context) error {
	if err := ms.strg.Ping(ctx); err != nil {
		return fmt.Errorf("DB is not connected: %w", err)
	}
	return nil
}

func (ms *MetricService) PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
	gs := make(map[string]float64)
	cs := make(map[string]int64)

	for _, gauge := range gauges {
		gs[gauge.Name] = gauge.Value
	}
	for _, counter := range counters {
		cs[counter.Name] += counter.Value
	}

	if err := ms.strg.WriteGaugesCounters(ctx, gs, cs); err != nil {
		return fmt.Errorf("failed to write gauges and counters: %w", err)
	}
	return nil
}
