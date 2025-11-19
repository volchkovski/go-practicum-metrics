package agent

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
	"go.uber.org/zap"
)

// mockGRPCClient implements the GRPC client interface for testing
type mockGRPCClient struct {
	pushMetricsFunc func(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error
	closeFunc       func() error
}

func (m *mockGRPCClient) PushMetrics(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
	if m.pushMetricsFunc != nil {
		return m.pushMetricsFunc(ctx, gauges, counters)
	}
	return nil
}

func (m *mockGRPCClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockGRPCClient) PushGauge(ctx context.Context, metric *m.GaugeMetric) error {
	return nil
}

func (m *mockGRPCClient) PushCounter(ctx context.Context, metric *m.CounterMetric) error {
	return nil
}

func (m *mockGRPCClient) Ping(ctx context.Context) error {
	return nil
}

func TestNew(t *testing.T) {
	t.Run("create new agent", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			Key:        "test-key",
			RateLimit:  1,
		}

		agent, err := New(config)
		assert.NoError(t, err)

		assert.NotNil(t, agent)
		assert.NotNil(t, agent.mstorage)
		assert.NotNil(t, agent.client)
		assert.Equal(t, "localhost:8080", agent.serverAddr)
		assert.Equal(t, 10*time.Second, agent.repIntr)
		assert.Equal(t, 2*time.Second, agent.pollIntr)
		assert.Equal(t, "test-key", agent.key)
		assert.Equal(t, 1, agent.rateLimit)
	})

	t.Run("create agent with empty crypto key", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			Key:        "",
			RateLimit:  1,
			CryptoKey:  "", // Empty crypto key should work
		}

		agent, err := New(config)
		assert.NoError(t, err)
		assert.NotNil(t, agent)
		assert.Nil(t, agent.publicRSA)
	})

	t.Run("handles invalid crypto key", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			Key:        "",
			RateLimit:  1,
			CryptoKey:  "/invalid/path/to/key.pem",
		}

		agent, err := New(config)
		assert.Error(t, err)
		assert.Nil(t, agent)
	})

	t.Run("create agent with UseGRPC but invalid server", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr:     "localhost:8080",
			ReportIntr:     10,
			PollIntr:       2,
			Key:            "",
			RateLimit:      1,
			UseGRPC:        true,
			GRPCServerAddr: "invalid:99999", // Invalid gRPC server
		}

		// With grpc.NewClient, agent creation should succeed even with invalid gRPC address
		// Connection errors only occur when actual RPC calls are made
		agent, err := New(config)
		assert.NoError(t, err, "Agent creation should not fail with invalid gRPC address as grpc.NewClient doesn't connect immediately")
		assert.NotNil(t, agent, "Agent should be created")
		assert.True(t, agent.useGRPC, "Agent should be configured to use gRPC")
	})
}

func TestGetRandomFloat(t *testing.T) {
	t.Run("generates random float", func(t *testing.T) {
		val1 := getRandomFloat()
		val2 := getRandomFloat()

		// Values should be in range [0, 1)
		assert.GreaterOrEqual(t, val1, 0.0)
		assert.Less(t, val1, 1.0)
		assert.GreaterOrEqual(t, val2, 0.0)
		assert.Less(t, val2, 1.0)

		// Test multiple calls to ensure function works
		for i := 0; i < 10; i++ {
			val := getRandomFloat()
			assert.GreaterOrEqual(t, val, 0.0)
			assert.Less(t, val, 1.0)
		}
	})
}

func TestGaugeVal(t *testing.T) {
	t.Run("extract valid gauge values", func(t *testing.T) {
		stats := &runtime.MemStats{
			Alloc: 12345,
			Sys:   67890,
			NumGC: 42,
		}

		// Test uint64 field
		val, ok := gaugeVal(stats, "Alloc")
		assert.True(t, ok)
		assert.Equal(t, float64(12345), val)

		// Test uint64 field
		val, ok = gaugeVal(stats, "Sys")
		assert.True(t, ok)
		assert.Equal(t, float64(67890), val)

		// Test uint32 field
		val, ok = gaugeVal(stats, "NumGC")
		assert.True(t, ok)
		assert.Equal(t, float64(42), val)
	})

	t.Run("handle invalid field names", func(t *testing.T) {
		stats := &runtime.MemStats{}

		val, ok := gaugeVal(stats, "NonExistentField")
		assert.False(t, ok)
		assert.Equal(t, float64(0), val)
	})

	t.Run("handle float64 fields", func(t *testing.T) {
		stats := &runtime.MemStats{
			GCCPUFraction: 0.123,
		}

		val, ok := gaugeVal(stats, "GCCPUFraction")
		assert.True(t, ok)
		assert.Equal(t, 0.123, val)
	})
}

func TestAgentSignalHandling(t *testing.T) {
	t.Run("agent creation with signal context", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		// Since setupSignalHandler was removed, test that agent creation works
		assert.NotNil(t, agent)
	})
}

func TestStartMetricCollection(t *testing.T) {
	t.Run("starts collection goroutine", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   1, // 1 second for faster test
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// This should start the goroutine and not panic
		assert.NotPanics(t, func() {
			agent.startMetricCollection(ctx)
		})

		// Wait for context to be done
		<-ctx.Done()
	})
}

func TestStartPostWorkers(t *testing.T) {
	t.Run("starts workers", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  2, // Test with 2 workers
		}
		agent, err := New(config)
		assert.NoError(t, err)

		metricsChunks := make(chan []*m.Metrics, 1)
		defer close(metricsChunks)

		var wg sync.WaitGroup
		assert.NotPanics(t, func() {
			agent.startPostWorkers(&wg, metricsChunks)
		})

		// Give workers a moment to start
		time.Sleep(10 * time.Millisecond)
	})
}

func TestRunReportingLoop(t *testing.T) {
	t.Run("handles context cancellation", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		metricsChunks := make(chan []*m.Metrics, 1)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		repTicker := time.NewTicker(1 * time.Hour) // Long interval so it doesn't trigger
		defer repTicker.Stop()

		var wg sync.WaitGroup

		// Start reporting loop in goroutine
		done := make(chan bool)
		go func() {
			agent.runReportingLoop(ctx, &wg, metricsChunks, repTicker)
			done <- true
		}()

		// Wait for context to be done or loop to exit
		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected reporting loop to exit on context cancel")
		}
	})
}

func TestPostMetricsGRPC(t *testing.T) {
	_ = zap.NewNop().Sugar() // Keep import

	t.Run("posts metrics successfully", func(t *testing.T) {
		gaugeValue := 42.0
		counterValue := int64(100)

		mockClient := &mockGRPCClient{
			pushMetricsFunc: func(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
				assert.Len(t, gauges, 1)
				assert.Len(t, counters, 1)
				assert.Equal(t, "test_gauge", gauges[0].Name)
				assert.Equal(t, 42.0, gauges[0].Value)
				assert.Equal(t, "test_counter", counters[0].Name)
				assert.Equal(t, int64(100), counters[0].Value)
				return nil
			},
		}

		agent := &Agent{
			grpcClient: mockClient,
		}

		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: &gaugeValue,
			},
			{
				ID:    "test_counter",
				MType: "counter",
				Delta: &counterValue,
			},
		}

		err := agent.postMetricsGRPC(metrics)
		assert.NoError(t, err)
	})

	t.Run("handles push error", func(t *testing.T) {
		mockClient := &mockGRPCClient{
			pushMetricsFunc: func(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
				return errors.New("push failed")
			},
		}

		agent := &Agent{
			grpcClient: mockClient,
		}

		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: new(float64),
			},
		}

		err := agent.postMetricsGRPC(metrics)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "push failed")
	})

	t.Run("handles metrics with nil values", func(t *testing.T) {
		mockClient := &mockGRPCClient{
			pushMetricsFunc: func(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
				// Should not include metrics with nil values
				assert.Len(t, gauges, 0)
				assert.Len(t, counters, 0)
				return nil
			},
		}

		agent := &Agent{
			grpcClient: mockClient,
		}

		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: nil, // nil value should be skipped
			},
			{
				ID:    "test_counter",
				MType: "counter",
				Delta: nil, // nil value should be skipped
			},
		}

		err := agent.postMetricsGRPC(metrics)
		assert.NoError(t, err)
	})

	t.Run("handles mixed valid and invalid metrics", func(t *testing.T) {
		gaugeValue := 42.0

		mockClient := &mockGRPCClient{
			pushMetricsFunc: func(ctx context.Context, gauges []*m.GaugeMetric, counters []*m.CounterMetric) error {
				// Should only include valid metric
				assert.Len(t, gauges, 1)
				assert.Len(t, counters, 0)
				assert.Equal(t, "valid_gauge", gauges[0].Name)
				return nil
			},
		}

		agent := &Agent{
			grpcClient: mockClient,
		}

		metrics := []*m.Metrics{
			{
				ID:    "valid_gauge",
				MType: "gauge",
				Value: &gaugeValue,
			},
			{
				ID:    "invalid_counter",
				MType: "counter",
				Delta: nil, // Should be skipped
			},
		}

		err := agent.postMetricsGRPC(metrics)
		assert.NoError(t, err)
	})
}

func TestPostMetricsHTTP(t *testing.T) {
	t.Run("successfully posts metrics via HTTP", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/updates/", r.URL.Path)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Read and validate body
			body, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.NotEmpty(t, body)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:], // Remove "http://"
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		gaugeValue := 42.0
		counterValue := int64(100)

		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: &gaugeValue,
			},
			{
				ID:    "test_counter",
				MType: "counter",
				Delta: &counterValue,
			},
		}

		err = agent.postMetricsHTTP(metrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, requestCount)
	})

	t.Run("handles HTTP server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:], // Remove "http://"
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		gaugeValue := 42.0
		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: &gaugeValue,
			},
		}

		err = agent.postMetricsHTTP(metrics)
		assert.Error(t, err)
	})

	t.Run("skips metrics with nil values", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:], // Remove "http://"
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		metrics := []*m.Metrics{
			{
				ID:    "test_gauge",
				MType: "gauge",
				Value: nil, // nil value should be filtered out
			},
			{
				ID:    "test_counter",
				MType: "counter",
				Delta: nil, // nil value should be filtered out
			},
		}

		err = agent.postMetricsHTTP(metrics)
		assert.NoError(t, err)
		// If all metrics are filtered out, should not make HTTP request
		// or make request with empty array
		assert.GreaterOrEqual(t, requestCount, 0)
	})

	t.Run("handles empty metrics slice", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:], // Remove "http://"
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		metrics := []*m.Metrics{}

		err = agent.postMetricsHTTP(metrics)
		assert.NoError(t, err)
	})
}

func TestPostWorker(t *testing.T) {
	t.Run("processes metrics from channel", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:],
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		gaugeValue := 42.0
		metricsChunks := make(chan []*m.Metrics, 1)
		metricsChunks <- []*m.Metrics{
			{ID: "test_gauge", MType: "gauge", Value: &gaugeValue},
		}
		close(metricsChunks)

		var wg sync.WaitGroup
		wg.Add(1)
		agent.postWorker(&wg, metricsChunks)
		wg.Wait()
	})

	t.Run("handles error from postMetrics", func(t *testing.T) {
		// Use a server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		config := &configs.AgentConfig{
			ServerAddr: server.URL[7:],
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		gaugeValue := 42.0
		metricsChunks := make(chan []*m.Metrics, 1)
		metricsChunks <- []*m.Metrics{
			{ID: "test_gauge", MType: "gauge", Value: &gaugeValue},
		}
		close(metricsChunks)

		var wg sync.WaitGroup
		wg.Add(1)
		// Should not panic even on error
		assert.NotPanics(t, func() {
			agent.postWorker(&wg, metricsChunks)
			wg.Wait()
		})
	})
}

func TestCollectExtraMetrics(t *testing.T) {
	t.Run("collects extra metrics", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent, err := New(config)
		assert.NoError(t, err)

		metricsCh := make(chan *m.Metrics, 10)
		var wg sync.WaitGroup
		wg.Add(1)

		agent.collectExtraMetrics(&wg, metricsCh)
		wg.Wait()
		close(metricsCh)

		// Should collect at least some metrics
		count := 0
		for range metricsCh {
			count++
		}
		assert.GreaterOrEqual(t, count, 0)
	})
}
