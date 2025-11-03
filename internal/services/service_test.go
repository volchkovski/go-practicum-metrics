package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/mock/gomock"
)

func TestMetricService(t *testing.T) {
	ctrl := gomock.NewController(t)
	strg := NewMockMetricStorage(ctrl)

	mservice := NewMetricService(strg)

	ctx := context.Background()

	t.Run("get gauge metric", func(t *testing.T) {
		strg.EXPECT().ReadGauge(ctx, "test").Return(float64(123), nil)
		m, err := mservice.GetGaugeMetric(ctx, "test")
		require.Nil(t, err)
		require.NotNil(t, m)
		assert.Equal(t, models.GaugeMetric{Name: "test", Value: float64(123)}, *m)
	})

	t.Run("get counter metric", func(t *testing.T) {
		strg.EXPECT().ReadCounter(ctx, "test").Return(int64(123), nil)
		m, err := mservice.GetCounterMetric(ctx, "test")
		require.Nil(t, err)
		require.NotNil(t, m)
		assert.Equal(t, models.CounterMetric{Name: "test", Value: int64(123)}, *m)
	})

	t.Run("push gauge metric", func(t *testing.T) {
		strg.EXPECT().WriteGauge(ctx, "test", float64(123)).Return(nil)
		err := mservice.PushGaugeMetric(ctx, &models.GaugeMetric{Name: "test", Value: float64(123)})
		require.Nil(t, err)
	})

	t.Run("push counter metric", func(t *testing.T) {
		strg.EXPECT().WriteCounter(ctx, "test", int64(123)).Return(nil)
		err := mservice.PushCounterMetric(ctx, &models.CounterMetric{Name: "test", Value: int64(123)})
		require.Nil(t, err)
	})

	t.Run("get all gauge metrics", func(t *testing.T) {
		strg.EXPECT().ReadAllGauges(ctx).Return(map[string]float64{"test": 123}, nil)
		ms, err := mservice.GetAllGaugeMetrics(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, ms)
		var gauges []models.GaugeMetric
		for _, m := range ms {
			gauges = append(gauges, *m)
		}
		assert.Equal(t, []models.GaugeMetric{{Name: "test", Value: 123}}, gauges)
	})

	t.Run("get all counter metrics", func(t *testing.T) {
		strg.EXPECT().ReadAllCounters(ctx).Return(map[string]int64{"test": 123}, nil)
		ms, err := mservice.GetAllCounterMetrics(ctx)
		require.Nil(t, err)
		require.NotEmpty(t, ms)
		var counters []models.CounterMetric
		for _, m := range ms {
			counters = append(counters, *m)
		}
		assert.Equal(t, []models.CounterMetric{{Name: "test", Value: 123}}, counters)
	})

	t.Run("close service", func(t *testing.T) {
		strg.EXPECT().Close().Return(nil)
		err := mservice.Close()
		require.Nil(t, err)
	})

	t.Run("ping database", func(t *testing.T) {
		strg.EXPECT().Ping(ctx).Return(nil)
		err := mservice.PingDB(ctx)
		require.Nil(t, err)
	})

	t.Run("push metrics batch", func(t *testing.T) {
		gauges := []*models.GaugeMetric{
			{Name: "gauge1", Value: 1.5},
			{Name: "gauge2", Value: 2.5},
		}
		counters := []*models.CounterMetric{
			{Name: "counter1", Value: 10},
			{Name: "counter2", Value: 20},
		}

		expectedGauges := map[string]float64{
			"gauge1": 1.5,
			"gauge2": 2.5,
		}
		expectedCounters := map[string]int64{
			"counter1": 10,
			"counter2": 20,
		}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(nil)
		err := mservice.PushMetrics(ctx, gauges, counters)
		require.Nil(t, err)
	})

	t.Run("push metrics with duplicate counters", func(t *testing.T) {
		counters := []*models.CounterMetric{
			{Name: "counter1", Value: 10},
			{Name: "counter1", Value: 5}, // Same name, should accumulate
		}

		expectedGauges := map[string]float64{}
		expectedCounters := map[string]int64{
			"counter1": 15, // 10 + 5
		}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(nil)
		err := mservice.PushMetrics(ctx, nil, counters)
		require.Nil(t, err)
	})
}
