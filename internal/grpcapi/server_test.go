package grpcapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/volchkovski/go-practicum-metrics/internal/grpcapi/pb"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

// MockMetricService is a mock implementation of MetricService
type MockMetricService struct {
	mock.Mock
}

func (ms *MockMetricService) PushGaugeMetric(ctx context.Context, metric *m.GaugeMetric) error {
	args := ms.Called(ctx, metric)
	return args.Error(0)
}

func (ms *MockMetricService) PushCounterMetric(ctx context.Context, metric *m.CounterMetric) error {
	args := ms.Called(ctx, metric)
	return args.Error(0)
}

func (ms *MockMetricService) PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
	args := ms.Called(ctx, gauges, counters)
	return args.Error(0)
}

func (ms *MockMetricService) GetGaugeMetric(ctx context.Context, name string) (*m.GaugeMetric, error) {
	args := ms.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*m.GaugeMetric), args.Error(1)
}

func (ms *MockMetricService) GetCounterMetric(ctx context.Context, name string) (*m.CounterMetric, error) {
	args := ms.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*m.CounterMetric), args.Error(1)
}

func (ms *MockMetricService) GetAllGaugeMetrics(ctx context.Context) ([]*m.GaugeMetric, error) {
	args := ms.Called(ctx)
	return args.Get(0).([]*m.GaugeMetric), args.Error(1)
}

func (ms *MockMetricService) GetAllCounterMetrics(ctx context.Context) ([]*m.CounterMetric, error) {
	args := ms.Called(ctx)
	return args.Get(0).([]*m.CounterMetric), args.Error(1)
}

func (ms *MockMetricService) PingDB(ctx context.Context) error {
	args := ms.Called(ctx)
	return args.Error(0)
}

func TestMetricsServer_PushGauge(t *testing.T) {
	mockService := new(MockMetricService)
	server := NewMetricsServer(mockService)

	testMetric := &m.GaugeMetric{
		Name:  "test_gauge",
		Value: 42.5,
	}

	mockService.On("PushGaugeMetric", mock.Anything, testMetric).Return(nil)

	req := &pb.PushGaugeRequest{
		Metric: &pb.GaugeMetric{
			Name:  "test_gauge",
			Value: 42.5,
		},
	}

	resp, err := server.PushGauge(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	mockService.AssertExpectations(t)
}

func TestMetricsServer_PushCounter(t *testing.T) {
	mockService := new(MockMetricService)
	server := NewMetricsServer(mockService)

	testMetric := &m.CounterMetric{
		Name:  "test_counter",
		Value: 100,
	}

	mockService.On("PushCounterMetric", mock.Anything, testMetric).Return(nil)

	req := &pb.PushCounterRequest{
		Metric: &pb.CounterMetric{
			Name:  "test_counter",
			Value: 100,
		},
	}

	resp, err := server.PushCounter(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	mockService.AssertExpectations(t)
}

func TestMetricsServer_GetGauge(t *testing.T) {
	mockService := new(MockMetricService)
	server := NewMetricsServer(mockService)

	expectedMetric := &m.GaugeMetric{
		Name:  "test_gauge",
		Value: 42.5,
	}

	mockService.On("GetGaugeMetric", mock.Anything, "test_gauge").Return(expectedMetric, nil)

	req := &pb.GetGaugeRequest{
		Name: "test_gauge",
	}

	resp, err := server.GetGauge(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test_gauge", resp.Metric.Name)
	assert.Equal(t, 42.5, resp.Metric.Value)
	mockService.AssertExpectations(t)
}

func TestMetricsServer_Ping(t *testing.T) {
	mockService := new(MockMetricService)
	server := NewMetricsServer(mockService)

	mockService.On("PingDB", mock.Anything).Return(nil)

	req := &pb.PingRequest{}

	resp, err := server.Ping(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Healthy)
	mockService.AssertExpectations(t)
}

func TestMetricsServer_PushMetrics(t *testing.T) {
	mockService := new(MockMetricService)
	server := NewMetricsServer(mockService)

	gauges := []*m.GaugeMetric{
		{Name: "gauge1", Value: 1.1},
		{Name: "gauge2", Value: 2.2},
	}

	counters := []*m.CounterMetric{
		{Name: "counter1", Value: 10},
		{Name: "counter2", Value: 20},
	}

	mockService.On("PushMetrics", mock.Anything, gauges, counters).Return(nil)

	req := &pb.PushMetricsRequest{
		Gauges: []*pb.GaugeMetric{
			{Name: "gauge1", Value: 1.1},
			{Name: "gauge2", Value: 2.2},
		},
		Counters: []*pb.CounterMetric{
			{Name: "counter1", Value: 10},
			{Name: "counter2", Value: 20},
		},
	}

	resp, err := server.PushMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.Success)
	mockService.AssertExpectations(t)
}

func TestMetricsServer_GetCounter(t *testing.T) {
	t.Run("gets counter successfully", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		expectedMetric := &m.CounterMetric{
			Name:  "test_counter",
			Value: 100,
		}

		mockService.On("GetCounterMetric", mock.Anything, "test_counter").Return(expectedMetric, nil)

		req := &pb.GetCounterRequest{
			Name: "test_counter",
		}

		resp, err := server.GetCounter(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test_counter", resp.Metric.Name)
		assert.Equal(t, int64(100), resp.Metric.Value)
		mockService.AssertExpectations(t)
	})

	t.Run("handles error getting counter", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		mockService.On("GetCounterMetric", mock.Anything, "missing_counter").Return(nil, assert.AnError)

		req := &pb.GetCounterRequest{
			Name: "missing_counter",
		}

		resp, err := server.GetCounter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

func TestMetricsServer_GetAllMetrics(t *testing.T) {
	t.Run("gets all metrics successfully", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		gauges := []*m.GaugeMetric{
			{Name: "gauge1", Value: 1.1},
			{Name: "gauge2", Value: 2.2},
		}

		counters := []*m.CounterMetric{
			{Name: "counter1", Value: 10},
			{Name: "counter2", Value: 20},
		}

		mockService.On("GetAllGaugeMetrics", mock.Anything).Return(gauges, nil)
		mockService.On("GetAllCounterMetrics", mock.Anything).Return(counters, nil)

		req := &pb.GetAllMetricsRequest{}

		resp, err := server.GetAllMetrics(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Gauges, 2)
		assert.Len(t, resp.Counters, 2)
		assert.Equal(t, "gauge1", resp.Gauges[0].Name)
		assert.Equal(t, 1.1, resp.Gauges[0].Value)
		assert.Equal(t, "counter1", resp.Counters[0].Name)
		assert.Equal(t, int64(10), resp.Counters[0].Value)
		mockService.AssertExpectations(t)
	})

	t.Run("handles error getting gauges", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		mockService.On("GetAllGaugeMetrics", mock.Anything).Return([]*m.GaugeMetric{}, assert.AnError)

		req := &pb.GetAllMetricsRequest{}

		resp, err := server.GetAllMetrics(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("handles error getting counters", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		gauges := []*m.GaugeMetric{
			{Name: "gauge1", Value: 1.1},
		}

		mockService.On("GetAllGaugeMetrics", mock.Anything).Return(gauges, nil)
		mockService.On("GetAllCounterMetrics", mock.Anything).Return([]*m.CounterMetric{}, assert.AnError)

		req := &pb.GetAllMetricsRequest{}

		resp, err := server.GetAllMetrics(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})
}

// Error handling tests
func TestMetricsServer_ErrorHandling(t *testing.T) {
	t.Run("PushGauge handles error", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		testMetric := &m.GaugeMetric{
			Name:  "test_gauge",
			Value: 42.5,
		}

		mockService.On("PushGaugeMetric", mock.Anything, testMetric).Return(assert.AnError)

		req := &pb.PushGaugeRequest{
			Metric: &pb.GaugeMetric{
				Name:  "test_gauge",
				Value: 42.5,
			},
		}

		resp, err := server.PushGauge(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("PushCounter handles error", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		testMetric := &m.CounterMetric{
			Name:  "test_counter",
			Value: 100,
		}

		mockService.On("PushCounterMetric", mock.Anything, testMetric).Return(assert.AnError)

		req := &pb.PushCounterRequest{
			Metric: &pb.CounterMetric{
				Name:  "test_counter",
				Value: 100,
			},
		}

		resp, err := server.PushCounter(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("PushMetrics handles error", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		gauges := []*m.GaugeMetric{{Name: "gauge1", Value: 1.1}}
		counters := []*m.CounterMetric{{Name: "counter1", Value: 10}}

		mockService.On("PushMetrics", mock.Anything, gauges, counters).Return(assert.AnError)

		req := &pb.PushMetricsRequest{
			Gauges:   []*pb.GaugeMetric{{Name: "gauge1", Value: 1.1}},
			Counters: []*pb.CounterMetric{{Name: "counter1", Value: 10}},
		}

		resp, err := server.PushMetrics(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("GetGauge handles error", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		mockService.On("GetGaugeMetric", mock.Anything, "missing_gauge").Return(nil, assert.AnError)

		req := &pb.GetGaugeRequest{
			Name: "missing_gauge",
		}

		resp, err := server.GetGauge(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		mockService.AssertExpectations(t)
	})

	t.Run("Ping handles error", func(t *testing.T) {
		mockService := new(MockMetricService)
		server := NewMetricsServer(mockService)

		mockService.On("PingDB", mock.Anything).Return(assert.AnError)

		req := &pb.PingRequest{}

		resp, err := server.Ping(context.Background(), req)

		// Ping returns Healthy:false instead of error
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.False(t, resp.Healthy)
		mockService.AssertExpectations(t)
	})
}

