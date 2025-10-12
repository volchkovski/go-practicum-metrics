package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGaugeMetric(t *testing.T) {
	t.Run("create gauge metric", func(t *testing.T) {
		metric := GaugeMetric{
			Name:  "test_gauge",
			Value: 123.45,
		}

		assert.Equal(t, "test_gauge", metric.Name)
		assert.Equal(t, 123.45, metric.Value)
	})

	t.Run("zero value gauge", func(t *testing.T) {
		metric := GaugeMetric{}

		assert.Empty(t, metric.Name)
		assert.Equal(t, float64(0), metric.Value)
	})
}

func TestCounterMetric(t *testing.T) {
	t.Run("create counter metric", func(t *testing.T) {
		metric := CounterMetric{
			Name:  "test_counter",
			Value: 42,
		}

		assert.Equal(t, "test_counter", metric.Name)
		assert.Equal(t, int64(42), metric.Value)
	})

	t.Run("zero value counter", func(t *testing.T) {
		metric := CounterMetric{}

		assert.Empty(t, metric.Name)
		assert.Equal(t, int64(0), metric.Value)
	})
}

func TestMetrics(t *testing.T) {
	t.Run("create metrics struct", func(t *testing.T) {
		gauge := &GaugeMetric{Name: "gauge1", Value: 1.5}
		counter := &CounterMetric{Name: "counter1", Value: 10}

		metrics := Metrics{
			ID:    "metric_id",
			MType: "gauge",
			Delta: &counter.Value,
			Value: &gauge.Value,
		}

		assert.Equal(t, "metric_id", metrics.ID)
		assert.Equal(t, "gauge", metrics.MType)
		assert.NotNil(t, metrics.Delta)
		assert.NotNil(t, metrics.Value)
		assert.Equal(t, int64(10), *metrics.Delta)
		assert.Equal(t, 1.5, *metrics.Value)
	})

	t.Run("nil pointers", func(t *testing.T) {
		metrics := Metrics{
			ID:    "test",
			MType: "counter",
		}

		assert.Nil(t, metrics.Delta)
		assert.Nil(t, metrics.Value)
	})
}
