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

func TestMetricService_ErrorHandling(t *testing.T) {
	ctrl := gomock.NewController(t)
	strg := NewMockMetricStorage(ctrl)
	mservice := NewMetricService(strg)
	ctx := context.Background()

	t.Run("get gauge metric returns error", func(t *testing.T) {
		strg.EXPECT().ReadGauge(ctx, "test").Return(float64(0), assert.AnError)
		m, err := mservice.GetGaugeMetric(ctx, "test")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "failed to get gauge metric")
	})

	t.Run("get counter metric returns error", func(t *testing.T) {
		strg.EXPECT().ReadCounter(ctx, "test").Return(int64(0), assert.AnError)
		m, err := mservice.GetCounterMetric(ctx, "test")
		assert.Error(t, err)
		assert.Nil(t, m)
		assert.Contains(t, err.Error(), "failed to get counter metric")
	})

	t.Run("push gauge metric returns error", func(t *testing.T) {
		strg.EXPECT().WriteGauge(ctx, "test", float64(123)).Return(assert.AnError)
		err := mservice.PushGaugeMetric(ctx, &models.GaugeMetric{Name: "test", Value: float64(123)})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push gauge metric")
	})

	t.Run("push counter metric returns error", func(t *testing.T) {
		strg.EXPECT().WriteCounter(ctx, "test", int64(123)).Return(assert.AnError)
		err := mservice.PushCounterMetric(ctx, &models.CounterMetric{Name: "test", Value: int64(123)})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to push counter metric")
	})

	t.Run("get all gauge metrics returns error", func(t *testing.T) {
		strg.EXPECT().ReadAllGauges(ctx).Return(nil, assert.AnError)
		ms, err := mservice.GetAllGaugeMetrics(ctx)
		assert.Error(t, err)
		assert.Nil(t, ms)
		assert.Contains(t, err.Error(), "failed to get all gauge metrics")
	})

	t.Run("get all counter metrics returns error", func(t *testing.T) {
		strg.EXPECT().ReadAllCounters(ctx).Return(nil, assert.AnError)
		ms, err := mservice.GetAllCounterMetrics(ctx)
		assert.Error(t, err)
		assert.Nil(t, ms)
		assert.Contains(t, err.Error(), "failed to get all counter metrics")
	})

	t.Run("ping database returns error", func(t *testing.T) {
		strg.EXPECT().Ping(ctx).Return(assert.AnError)
		err := mservice.PingDB(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DB is not connected")
	})

	t.Run("push metrics batch returns error", func(t *testing.T) {
		gauges := []*models.GaugeMetric{{Name: "gauge1", Value: 1.5}}
		counters := []*models.CounterMetric{{Name: "counter1", Value: 10}}

		expectedGauges := map[string]float64{"gauge1": 1.5}
		expectedCounters := map[string]int64{"counter1": 10}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(assert.AnError)
		err := mservice.PushMetrics(ctx, gauges, counters)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write gauges and counters")
	})

	t.Run("close service returns error", func(t *testing.T) {
		strg.EXPECT().Close().Return(assert.AnError)
		err := mservice.Close()
		assert.Error(t, err)
	})
}

func TestMetricService_EdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	strg := NewMockMetricStorage(ctrl)
	mservice := NewMetricService(strg)
	ctx := context.Background()

	t.Run("push metrics with empty lists", func(t *testing.T) {
		expectedGauges := map[string]float64{}
		expectedCounters := map[string]int64{}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(nil)
		err := mservice.PushMetrics(ctx, []*models.GaugeMetric{}, []*models.CounterMetric{})
		assert.NoError(t, err)
	})

	t.Run("push metrics with nil lists", func(t *testing.T) {
		expectedGauges := map[string]float64{}
		expectedCounters := map[string]int64{}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(nil)
		err := mservice.PushMetrics(ctx, nil, nil)
		assert.NoError(t, err)
	})

	t.Run("get all gauges with empty storage", func(t *testing.T) {
		strg.EXPECT().ReadAllGauges(ctx).Return(map[string]float64{}, nil)
		ms, err := mservice.GetAllGaugeMetrics(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, ms)
		assert.Empty(t, ms)
	})

	t.Run("get all counters with empty storage", func(t *testing.T) {
		strg.EXPECT().ReadAllCounters(ctx).Return(map[string]int64{}, nil)
		ms, err := mservice.GetAllCounterMetrics(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, ms)
		assert.Empty(t, ms)
	})

	t.Run("push metrics with duplicate gauges overwrites", func(t *testing.T) {
		gauges := []*models.GaugeMetric{
			{Name: "gauge1", Value: 1.5},
			{Name: "gauge1", Value: 2.5}, // Same name, last one wins
		}

		expectedGauges := map[string]float64{
			"gauge1": 2.5, // Last value
		}
		expectedCounters := map[string]int64{}

		strg.EXPECT().WriteGaugesCounters(ctx, expectedGauges, expectedCounters).Return(nil)
		err := mservice.PushMetrics(ctx, gauges, nil)
		assert.NoError(t, err)
	})

	t.Run("get all gauges with multiple metrics", func(t *testing.T) {
		strg.EXPECT().ReadAllGauges(ctx).Return(map[string]float64{
			"gauge1": 1.1,
			"gauge2": 2.2,
			"gauge3": 3.3,
		}, nil)
		ms, err := mservice.GetAllGaugeMetrics(ctx)
		assert.NoError(t, err)
		assert.Len(t, ms, 3)
	})

	t.Run("get all counters with multiple metrics", func(t *testing.T) {
		strg.EXPECT().ReadAllCounters(ctx).Return(map[string]int64{
			"counter1": 10,
			"counter2": 20,
			"counter3": 30,
		}, nil)
		ms, err := mservice.GetAllCounterMetrics(ctx)
		assert.NoError(t, err)
		assert.Len(t, ms, 3)
	})
}
