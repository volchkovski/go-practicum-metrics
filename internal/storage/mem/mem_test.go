package mem

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.NotNil(t, storage)
	assert.NotNil(t, storage.gauges)
	assert.NotNil(t, storage.counters)
	assert.Equal(t, 0, len(storage.gauges))
	assert.Equal(t, 0, len(storage.counters))
}

func TestMemStorage_Close(t *testing.T) {
	storage := NewMemStorage()
	err := storage.Close()
	assert.NoError(t, err)
}

func TestMemStorage_WriteGauge(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{"write positive value", "metric1", 123.45},
		{"write negative value", "metric2", -67.89},
		{"write zero value", "metric3", 0.0},
		{"overwrite existing value", "metric1", 999.99},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := storage.WriteGauge(ctx, tc.key, tc.value)
			require.NoError(t, err)

			// Check that value was stored correctly
			assert.Equal(t, tc.value, storage.gauges[tc.key])
		})
	}
}

func TestMemStorage_WriteGauge_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := storage.WriteGauge(ctx, "test", 123.45)
	assert.Equal(t, ErrCanceled, err)
}

func TestMemStorage_WriteCounter(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	tests := []struct {
		name          string
		key           string
		value         int64
		expectedTotal int64
	}{
		{"write first value", "counter1", 10, 10},
		{"add to existing counter", "counter1", 5, 15},
		{"add negative value", "counter1", -3, 12},
		{"new counter with zero", "counter2", 0, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := storage.WriteCounter(ctx, tc.key, tc.value)
			require.NoError(t, err)

			// Check that value was accumulated correctly
			assert.Equal(t, tc.expectedTotal, storage.counters[tc.key])
		})
	}
}

func TestMemStorage_WriteCounter_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := storage.WriteCounter(ctx, "test", 123)
	assert.Equal(t, ErrCanceled, err)
}

func TestMemStorage_ReadGauge(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	// Store test data
	storage.gauges["existing"] = 456.78

	tests := []struct {
		name        string
		key         string
		expectedVal float64
		expectError bool
	}{
		{"read existing gauge", "existing", 456.78, false},
		{"read non-existing gauge", "missing", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := storage.ReadGauge(ctx, tc.key)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestMemStorage_ReadGauge_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	val, err := storage.ReadGauge(ctx, "test")
	assert.Equal(t, ErrCanceled, err)
	assert.Equal(t, float64(0), val)
}

func TestMemStorage_ReadCounter(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	// Store test data
	storage.counters["existing"] = 789

	tests := []struct {
		name        string
		key         string
		expectedVal int64
		expectError bool
	}{
		{"read existing counter", "existing", 789, false},
		{"read non-existing counter", "missing", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := storage.ReadCounter(ctx, tc.key)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedVal, val)
			}
		})
	}
}

func TestMemStorage_ReadCounter_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	val, err := storage.ReadCounter(ctx, "test")
	assert.Equal(t, ErrCanceled, err)
	assert.Equal(t, int64(0), val)
}

func TestMemStorage_ReadAllGauges(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	// Store test data
	testData := map[string]float64{
		"gauge1": 1.1,
		"gauge2": 2.2,
		"gauge3": 3.3,
	}
	storage.gauges = testData

	gauges, err := storage.ReadAllGauges(ctx)
	require.NoError(t, err)
	assert.Equal(t, testData, gauges)
}

func TestMemStorage_ReadAllGauges_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	gauges, err := storage.ReadAllGauges(ctx)
	assert.Equal(t, ErrCanceled, err)
	assert.Nil(t, gauges)
}

func TestMemStorage_ReadAllCounters(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	// Store test data
	testData := map[string]int64{
		"counter1": 10,
		"counter2": 20,
		"counter3": 30,
	}
	storage.counters = testData

	counters, err := storage.ReadAllCounters(ctx)
	require.NoError(t, err)
	assert.Equal(t, testData, counters)
}

func TestMemStorage_ReadAllCounters_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	counters, err := storage.ReadAllCounters(ctx)
	assert.Equal(t, ErrCanceled, err)
	assert.Nil(t, counters)
}

func TestMemStorage_WriteGaugesCounters(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	// Prepare initial data
	storage.gauges["existing_gauge"] = 100.0
	storage.counters["existing_counter"] = 50

	gauges := map[string]float64{
		"gauge1":         1.1,
		"gauge2":         2.2,
		"existing_gauge": 999.9, // Should overwrite
	}

	counters := map[string]int64{
		"counter1":         10,
		"counter2":         20,
		"existing_counter": 25, // Should add to existing (50 + 25 = 75)
	}

	err := storage.WriteGaugesCounters(ctx, gauges, counters)
	require.NoError(t, err)

	// Check gauges were copied correctly (overwrite behavior)
	assert.Equal(t, float64(1.1), storage.gauges["gauge1"])
	assert.Equal(t, float64(2.2), storage.gauges["gauge2"])
	assert.Equal(t, float64(999.9), storage.gauges["existing_gauge"])

	// Check counters were added correctly (accumulate behavior)
	assert.Equal(t, int64(10), storage.counters["counter1"])
	assert.Equal(t, int64(20), storage.counters["counter2"])
	assert.Equal(t, int64(75), storage.counters["existing_counter"])
}

func TestMemStorage_WriteGaugesCounters_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	gauges := map[string]float64{"test": 1.0}
	counters := map[string]int64{"test": 1}

	err := storage.WriteGaugesCounters(ctx, gauges, counters)
	assert.Equal(t, ErrCanceled, err)
}

func TestMemStorage_Ping(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	err := storage.Ping(ctx)
	assert.NoError(t, err)
}

func TestMemStorage_Ping_Canceled(t *testing.T) {
	storage := NewMemStorage()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := storage.Ping(ctx)
	assert.Equal(t, ErrCanceled, err)
}

func TestMemStorage_ConcurrentAccess(t *testing.T) {
	storage := NewMemStorage()
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	// Test concurrent gauge operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("gauge_%d", id)
				value := float64(j)

				// Write
				err := storage.WriteGauge(ctx, key, value)
				assert.NoError(t, err)

				// Read
				_, _ = storage.ReadGauge(ctx, key)
				// May or may not exist due to concurrency, so don't assert error
			}
		}(i)
	}

	// Test concurrent counter operations
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("counter_%d", id)

				// Write
				err := storage.WriteCounter(ctx, key, 1)
				assert.NoError(t, err)

				// Read
				_, _ = storage.ReadCounter(ctx, key)
				// May or may not exist due to concurrency, so don't assert error
			}
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()

	// Verify some data was written
	allGauges, err := storage.ReadAllGauges(ctx)
	require.NoError(t, err)
	assert.True(t, len(allGauges) > 0)

	allCounters, err := storage.ReadAllCounters(ctx)
	require.NoError(t, err)
	assert.True(t, len(allCounters) > 0)
}

func TestMemStorage_ContextTimeout(t *testing.T) {
	storage := NewMemStorage()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to timeout
	time.Sleep(1 * time.Millisecond)

	// All operations should return ErrCanceled
	err := storage.WriteGauge(ctx, "test", 1.0)
	assert.Equal(t, ErrCanceled, err)

	err = storage.WriteCounter(ctx, "test", 1)
	assert.Equal(t, ErrCanceled, err)

	_, err = storage.ReadGauge(ctx, "test")
	assert.Equal(t, ErrCanceled, err)

	_, err = storage.ReadCounter(ctx, "test")
	assert.Equal(t, ErrCanceled, err)

	_, err = storage.ReadAllGauges(ctx)
	assert.Equal(t, ErrCanceled, err)

	_, err = storage.ReadAllCounters(ctx)
	assert.Equal(t, ErrCanceled, err)

	err = storage.WriteGaugesCounters(ctx, nil, nil)
	assert.Equal(t, ErrCanceled, err)

	err = storage.Ping(ctx)
	assert.Equal(t, ErrCanceled, err)
}
