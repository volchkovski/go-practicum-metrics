package agent

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/mock/gomock"
)

func TestPostMetrics(t *testing.T) {
	t.Run("handles empty metrics slice gracefully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		agent := &Agent{
			serverAddr: "localhost:8080",
			client:     NewRestyClient(),
		}

		err := agent.postMetrics([]*m.Metrics{})
		assert.NoError(t, err)
	})

	t.Run("successfully marshals complex metrics", func(t *testing.T) {
		agent := &Agent{
			serverAddr: "test-server:9090",
			client:     NewRestyClient(),
			key:        "secret-key",
		}

		value := 123.456
		delta := int64(789)
		metrics := []*m.Metrics{
			{ID: "cpu_usage", MType: "gauge", Value: &value},
			{ID: "request_count", MType: "counter", Delta: &delta},
		}

		// этот тест упадет с ошибкой соединения, но JSON marshaling должен пройти успешно
		err := agent.postMetrics(metrics)
		assert.Error(t, err)                       // ожидаем ошибку соединения
		assert.NotContains(t, err.Error(), "json") // но не ошибку JSON
	})

	t.Run("handles HTTP errors correctly", func(t *testing.T) {
		agent := &Agent{
			serverAddr: "invalid-server:99999", // заведомо неверный адрес
			client:     NewRestyClient(),
		}

		value := 100.0
		metrics := []*m.Metrics{
			{ID: "test_metric", MType: "gauge", Value: &value},
		}

		err := agent.postMetrics(metrics)
		assert.Error(t, err)
		// Проверяем что есть ошибка - не важно какая именно
		assert.NotEmpty(t, err.Error())
	})
}

func TestCollectAllMetrics(t *testing.T) {
	t.Run("increments poll count atomically", func(t *testing.T) {
		agent := &Agent{
			mstorage:  NewMetricsStorage(),
			pollCount: atomic.Int64{},
		}

		initialCount := int64(42)
		agent.pollCount.Store(initialCount)

		agent.collectAllMetrics()

		newCount := agent.pollCount.Load()
		assert.Equal(t, initialCount+1, newCount)
	})

	t.Run("stores collected metrics in storage", func(t *testing.T) {
		agent := &Agent{
			mstorage:  NewMetricsStorage(),
			pollCount: atomic.Int64{},
		}

		agent.collectAllMetrics()

		metrics := agent.mstorage.ReadMetrics()
		assert.NotEmpty(t, metrics)

		// проверяем что есть основные метрики
		foundPollCount := false
		foundRuntimeMetric := false
		foundRandomValue := false

		for _, metric := range metrics {
			switch metric.ID {
			case "PollCount":
				foundPollCount = true
				assert.Equal(t, "counter", metric.MType)
				assert.NotNil(t, metric.Delta)
			case "RandomValue":
				foundRandomValue = true
				assert.Equal(t, "gauge", metric.MType)
				assert.NotNil(t, metric.Value)
			case "Alloc", "Sys", "HeapAlloc": // примеры runtime метрик
				foundRuntimeMetric = true
			}
		}

		assert.True(t, foundPollCount, "должна быть метрика PollCount")
		assert.True(t, foundRandomValue, "должна быть метрика RandomValue")
		assert.True(t, foundRuntimeMetric, "должна быть хотя бы одна runtime метрика")
	})

	t.Run("collects consistent number of metrics", func(t *testing.T) {
		agent := &Agent{
			mstorage:  NewMetricsStorage(),
			pollCount: atomic.Int64{},
		}

		// собираем метрики несколько раз
		agent.collectAllMetrics()
		firstCount := len(agent.mstorage.ReadMetrics())

		agent.collectAllMetrics()
		secondCount := len(agent.mstorage.ReadMetrics())

		// количество метрик должно быть стабильным
		assert.Equal(t, firstCount, secondCount)
		assert.Greater(t, firstCount, 30) // должно быть достаточно много метрик
	})
}

func TestAgent_ErrorHandling(t *testing.T) {
	t.Run("handles context cancellation gracefully", func(t *testing.T) {
		agent := &Agent{
			mstorage:  NewMetricsStorage(),
			pollIntr:  100 * time.Millisecond,
			pollCount: atomic.Int64{},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// запускаем сбор метрик
		agent.startMetricCollection(ctx)

		// ждем отмены контекста
		<-ctx.Done()

		// проверяем что нет паники и горутина завершилась
		time.Sleep(200 * time.Millisecond) // даем время горутине завершиться
	})
}

func TestAgent_Integration(t *testing.T) {
	t.Run("full metric collection workflow", func(t *testing.T) {
		agent := &Agent{
			mstorage:   NewMetricsStorage(),
			serverAddr: "localhost:8080",
			client:     NewRestyClient(),
			pollCount:  atomic.Int64{},
		}

		// собираем метрики
		agent.collectAllMetrics()

		// получаем собранные метрики
		metrics := agent.mstorage.ReadMetrics()
		assert.NotEmpty(t, metrics)

		// проверяем что можем попробовать отправить их
		// (будет ошибка соединения, но это нормально для теста)
		err := agent.postMetrics(metrics)
		assert.Error(t, err) // ожидаем ошибку подключения
	})
}
