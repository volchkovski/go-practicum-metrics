package agent

import (
	"context"
	"os"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	m "github.com/volchkovski/go-practicum-metrics/internal/models"
)

func TestNew(t *testing.T) {
	t.Run("create new agent", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			Key:        "test-key",
			RateLimit:  1,
		}

		agent := New(config)

		assert.NotNil(t, agent)
		assert.NotNil(t, agent.mstorage)
		assert.NotNil(t, agent.client)
		assert.Equal(t, "localhost:8080", agent.serverAddr)
		assert.Equal(t, 10*time.Second, agent.repIntr)
		assert.Equal(t, 2*time.Second, agent.pollIntr)
		assert.Equal(t, "test-key", agent.key)
		assert.Equal(t, 1, agent.rateLimit)
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

func TestSetupSignalHandler(t *testing.T) {
	t.Run("creates signal channel", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent := New(config)

		interrupt := agent.setupSignalHandler()

		assert.NotNil(t, interrupt)

		// Test that we can send a signal to the channel
		go func() {
			time.Sleep(10 * time.Millisecond)
			interrupt <- syscall.SIGTERM
		}()

		select {
		case sig := <-interrupt:
			assert.Equal(t, syscall.SIGTERM, sig)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected to receive signal")
		}
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
		agent := New(config)

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
		agent := New(config)

		metricsChunks := make(chan []*m.Metrics, 1)
		defer close(metricsChunks)

		assert.NotPanics(t, func() {
			agent.startPostWorkers(metricsChunks)
		})

		// Give workers a moment to start
		time.Sleep(10 * time.Millisecond)
	})
}

func TestRunReportingLoop(t *testing.T) {
	t.Run("handles interrupt signal", func(t *testing.T) {
		config := &configs.AgentConfig{
			ServerAddr: "localhost:8080",
			ReportIntr: 10,
			PollIntr:   2,
			RateLimit:  1,
		}
		agent := New(config)

		metricsChunks := make(chan []*m.Metrics, 1)
		defer close(metricsChunks)

		interrupt := make(chan os.Signal, 1)
		repTicker := time.NewTicker(1 * time.Hour) // Long interval so it doesn't trigger
		defer repTicker.Stop()

		// Start reporting loop in goroutine
		done := make(chan bool)
		go func() {
			agent.runReportingLoop(metricsChunks, interrupt, repTicker)
			done <- true
		}()

		// Send interrupt signal
		interrupt <- syscall.SIGTERM

		// Wait for loop to exit
		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Expected reporting loop to exit on interrupt")
		}
	})
}
