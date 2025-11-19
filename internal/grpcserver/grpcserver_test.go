package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/zap"
)

// MockMetricService implements grpcapi.MetricService for testing
type MockMetricService struct{}

func (m *MockMetricService) PushGaugeMetric(ctx context.Context, metric *models.GaugeMetric) error {
	return nil
}

func (m *MockMetricService) PushCounterMetric(ctx context.Context, metric *models.CounterMetric) error {
	return nil
}

func (m *MockMetricService) PushMetrics(ctx context.Context, gauges []*models.GaugeMetric, counters []*models.CounterMetric) error {
	return nil
}

func (m *MockMetricService) GetGaugeMetric(ctx context.Context, name string) (*models.GaugeMetric, error) {
	return &models.GaugeMetric{Name: name, Value: 42.0}, nil
}

func (m *MockMetricService) GetCounterMetric(ctx context.Context, name string) (*models.CounterMetric, error) {
	return &models.CounterMetric{Name: name, Value: 100}, nil
}

func (m *MockMetricService) GetAllGaugeMetrics(ctx context.Context) ([]*models.GaugeMetric, error) {
	return []*models.GaugeMetric{}, nil
}

func (m *MockMetricService) GetAllCounterMetrics(ctx context.Context) ([]*models.CounterMetric, error) {
	return []*models.CounterMetric{}, nil
}

func (m *MockMetricService) PingDB(ctx context.Context) error {
	return nil
}

func TestNew(t *testing.T) {
	logger := zap.NewNop().Sugar()
	service := &MockMetricService{}

	t.Run("creates server without trusted subnet", func(t *testing.T) {
		server := New("localhost:0", service, logger, "")
		
		assert.NotNil(t, server)
		assert.Equal(t, "localhost:0", server.address)
		assert.NotNil(t, server.server)
		assert.NotNil(t, server.notify)
		assert.Equal(t, logger, server.logger)
	})

	t.Run("creates server with trusted subnet", func(t *testing.T) {
		server := New("localhost:0", service, logger, "192.168.1.0/24")
		
		assert.NotNil(t, server)
		assert.Equal(t, "localhost:0", server.address)
		assert.NotNil(t, server.server)
	})
}

func TestGRPCServer_Start_And_Shutdown(t *testing.T) {
	logger := zap.NewNop().Sugar()
	service := &MockMetricService{}

	t.Run("starts and shuts down successfully", func(t *testing.T) {
		// Use port 0 to get a random available port
		server := New("localhost:0", service, logger, "")
		
		// Start server
		server.Start()
		
		// Give it time to start
		time.Sleep(100 * time.Millisecond)
		
		// Verify listener is created
		assert.NotNil(t, server.listener)
		
		// Shutdown
		err := server.Shutdown()
		assert.NoError(t, err)
		
		// Wait for shutdown
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("returns error for invalid address", func(t *testing.T) {
		server := New("invalid:address:format", service, logger, "")
		
		server.Start()
		
		// Wait for error
		select {
		case err := <-server.Notify():
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to listen")
		case <-time.After(1 * time.Second):
			t.Fatal("expected error but got timeout")
		}
	})

	// Note: Full integration test with bufconn is complex and requires proper proto setup
	// This is covered by other integration tests
}

func TestGRPCServer_Notify(t *testing.T) {
	logger := zap.NewNop().Sugar()
	service := &MockMetricService{}

	t.Run("returns notify channel", func(t *testing.T) {
		server := New("localhost:0", service, logger, "")
		
		notify := server.Notify()
		assert.NotNil(t, notify)
		
		// Channel should be receive-only
		_, ok := interface{}(notify).(<-chan error)
		assert.True(t, ok)
	})
}
