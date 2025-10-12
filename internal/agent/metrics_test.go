package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeMetricNames(t *testing.T) {
	t.Run("runtime metric names exist", func(t *testing.T) {
		assert.NotEmpty(t, runtimeMetricNames)
		assert.Greater(t, len(runtimeMetricNames), 20)
	})

	t.Run("contains expected metrics", func(t *testing.T) {
		expectedMetrics := []string{
			"Alloc", "HeapAlloc", "HeapInuse", "TotalAlloc",
			"Mallocs", "Frees", "NumGC", "Sys",
		}

		for _, expected := range expectedMetrics {
			assert.Contains(t, runtimeMetricNames, expected)
		}
	})

	t.Run("no duplicate names", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, name := range runtimeMetricNames {
			assert.False(t, seen[name], "Duplicate metric name: %s", name)
			seen[name] = true
		}
	})
}
