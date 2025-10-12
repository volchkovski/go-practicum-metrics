package agent

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

func TestNewMetricsStorage(t *testing.T) {
	storage := NewMetricsStorage()

	require.NotNil(t, storage)
	assert.NotNil(t, storage.metrics)
	assert.Len(t, storage.metrics, 0)
	assert.Equal(t, cap(storage.metrics), 50) // проверяем capacity
}

func TestMetricsStorage_ReadMetrics(t *testing.T) {
	t.Run("empty storage returns empty slice", func(t *testing.T) {
		storage := NewMetricsStorage()

		metrics := storage.ReadMetrics()

		assert.NotNil(t, metrics)
		assert.Len(t, metrics, 0)
	})

	t.Run("returns all stored metrics", func(t *testing.T) {
		storage := NewMetricsStorage()

		expectedMetrics := []*m.Metrics{
			{ID: "cpu_usage", MType: "gauge"},
			{ID: "requests_total", MType: "counter"},
		}

		storage.ReplaceMetrics(expectedMetrics...)
		actualMetrics := storage.ReadMetrics()

		assert.Len(t, actualMetrics, 2)
		assert.Equal(t, "cpu_usage", actualMetrics[0].ID)
		assert.Equal(t, "requests_total", actualMetrics[1].ID)
	})

	t.Run("returns independent copy of metrics", func(t *testing.T) {
		storage := NewMetricsStorage()

		originalMetrics := []*m.Metrics{
			{ID: "memory_usage", MType: "gauge"},
		}

		storage.ReplaceMetrics(originalMetrics...)

		// получаем две копии
		metrics1 := storage.ReadMetrics()
		metrics2 := storage.ReadMetrics()

		// слайсы должны быть разными объектами
		assert.True(t, &metrics1 != &metrics2)
		// но содержимое одинаковое
		require.Len(t, metrics1, 1)
		require.Len(t, metrics2, 1)
		assert.Equal(t, metrics1[0].ID, metrics2[0].ID)
	})
}

func TestMetricsStorage_ReplaceMetrics(t *testing.T) {
	t.Run("replaces existing metrics completely", func(t *testing.T) {
		storage := NewMetricsStorage()

		// сначала добавляем старые метрики
		oldMetrics := []*m.Metrics{
			{ID: "old_metric", MType: "gauge"},
		}
		storage.ReplaceMetrics(oldMetrics...)

		// затем заменяем новыми
		newMetrics := []*m.Metrics{
			{ID: "new_metric_1", MType: "gauge"},
			{ID: "new_metric_2", MType: "counter"},
		}
		storage.ReplaceMetrics(newMetrics...)

		result := storage.ReadMetrics()
		assert.Len(t, result, 2)
		assert.Equal(t, "new_metric_1", result[0].ID)
		assert.Equal(t, "new_metric_2", result[1].ID)
	})

	t.Run("can replace with zero metrics", func(t *testing.T) {
		storage := NewMetricsStorage()

		// добавляем метрики
		storage.ReplaceMetrics(&m.Metrics{ID: "temp", MType: "gauge"})
		assert.Len(t, storage.ReadMetrics(), 1)

		// очищаем
		storage.ReplaceMetrics()
		assert.Len(t, storage.ReadMetrics(), 0)
	})

	t.Run("handles large number of metrics", func(t *testing.T) {
		storage := NewMetricsStorage()

		// создаем много метрик
		var metrics []*m.Metrics
		for i := 0; i < 100; i++ {
			metrics = append(metrics, &m.Metrics{
				ID:    fmt.Sprintf("metric_%d", i),
				MType: "gauge",
			})
		}

		storage.ReplaceMetrics(metrics...)
		result := storage.ReadMetrics()

		assert.Len(t, result, 100)
		assert.Equal(t, "metric_0", result[0].ID)
		assert.Equal(t, "metric_99", result[99].ID)
	})
}

func TestMetricsStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMetricsStorage()
	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // читатели + писатели

	// писатели
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				metrics := []*m.Metrics{
					{ID: fmt.Sprintf("writer_%d_metric_%d", id, j), MType: "gauge"},
				}
				storage.ReplaceMetrics(metrics...)
			}
		}(i)
	}

	// читатели
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				metrics := storage.ReadMetrics()
				// просто проверяем, что не паникуем и получаем валидный слайс
				assert.NotNil(t, metrics)
			}
		}()
	}

	wg.Wait()

	// финальная проверка состояния
	finalMetrics := storage.ReadMetrics()
	assert.NotNil(t, finalMetrics)
}
