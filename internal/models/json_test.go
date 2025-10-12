package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_JSONSerialization(t *testing.T) {
	t.Run("marshals gauge metric correctly", func(t *testing.T) {
		value := 123.456
		metric := Metrics{
			ID:    "cpu_usage",
			MType: "gauge",
			Value: &value,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		expected := `{"id":"cpu_usage","type":"gauge","value":123.456}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("marshals counter metric correctly", func(t *testing.T) {
		delta := int64(789)
		metric := Metrics{
			ID:    "request_count",
			MType: "counter",
			Delta: &delta,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		expected := `{"id":"request_count","type":"counter","delta":789}`
		assert.JSONEq(t, expected, string(data))
	})

	t.Run("marshals metric with both fields", func(t *testing.T) {
		value := 100.5
		delta := int64(42)
		metric := Metrics{
			ID:    "hybrid_metric",
			MType: "gauge",
			Value: &value,
			Delta: &delta,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		// Проверяем что оба поля присутствуют
		assert.Contains(t, string(data), `"value":100.5`)
		assert.Contains(t, string(data), `"delta":42`)
		assert.Contains(t, string(data), `"id":"hybrid_metric"`)
		assert.Contains(t, string(data), `"type":"gauge"`)
	})

	t.Run("marshals metric with nil values", func(t *testing.T) {
		metric := Metrics{
			ID:    "empty_metric",
			MType: "gauge",
			Value: nil,
			Delta: nil,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		expected := `{"id":"empty_metric","type":"gauge"}`
		assert.JSONEq(t, expected, string(data))
	})
}

func TestMetrics_JSONDeserialization(t *testing.T) {
	t.Run("unmarshals gauge metric correctly", func(t *testing.T) {
		jsonData := `{"id":"cpu_load","type":"gauge","value":67.89}`

		var metric Metrics
		err := json.Unmarshal([]byte(jsonData), &metric)
		require.NoError(t, err)

		assert.Equal(t, "cpu_load", metric.ID)
		assert.Equal(t, "gauge", metric.MType)
		require.NotNil(t, metric.Value)
		assert.Equal(t, 67.89, *metric.Value)
		assert.Nil(t, metric.Delta)
	})

	t.Run("unmarshals counter metric correctly", func(t *testing.T) {
		jsonData := `{"id":"total_requests","type":"counter","delta":1234}`

		var metric Metrics
		err := json.Unmarshal([]byte(jsonData), &metric)
		require.NoError(t, err)

		assert.Equal(t, "total_requests", metric.ID)
		assert.Equal(t, "counter", metric.MType)
		require.NotNil(t, metric.Delta)
		assert.Equal(t, int64(1234), *metric.Delta)
		assert.Nil(t, metric.Value)
	})

	t.Run("unmarshals metric with both fields", func(t *testing.T) {
		jsonData := `{"id":"test_metric","type":"gauge","value":99.9,"delta":555}`

		var metric Metrics
		err := json.Unmarshal([]byte(jsonData), &metric)
		require.NoError(t, err)

		assert.Equal(t, "test_metric", metric.ID)
		assert.Equal(t, "gauge", metric.MType)
		require.NotNil(t, metric.Value)
		require.NotNil(t, metric.Delta)
		assert.Equal(t, 99.9, *metric.Value)
		assert.Equal(t, int64(555), *metric.Delta)
	})

	t.Run("handles missing optional fields", func(t *testing.T) {
		jsonData := `{"id":"minimal_metric","type":"counter"}`

		var metric Metrics
		err := json.Unmarshal([]byte(jsonData), &metric)
		require.NoError(t, err)

		assert.Equal(t, "minimal_metric", metric.ID)
		assert.Equal(t, "counter", metric.MType)
		assert.Nil(t, metric.Value)
		assert.Nil(t, metric.Delta)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		invalidJSON := `{"id":"broken","type":"gauge","value":"not_a_number"}`

		var metric Metrics
		err := json.Unmarshal([]byte(invalidJSON), &metric)
		assert.Error(t, err)
	})
}

func TestMetrics_RoundTripSerialization(t *testing.T) {
	t.Run("gauge metric round trip", func(t *testing.T) {
		value := 3.14159
		original := Metrics{
			ID:    "pi_value",
			MType: "gauge",
			Value: &value,
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored Metrics
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		// Compare
		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.MType, restored.MType)
		require.NotNil(t, restored.Value)
		assert.Equal(t, *original.Value, *restored.Value)
		assert.Nil(t, restored.Delta)
	})

	t.Run("counter metric round trip", func(t *testing.T) {
		delta := int64(9999)
		original := Metrics{
			ID:    "error_count",
			MType: "counter",
			Delta: &delta,
		}

		// Marshal
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var restored Metrics
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		// Compare
		assert.Equal(t, original.ID, restored.ID)
		assert.Equal(t, original.MType, restored.MType)
		require.NotNil(t, restored.Delta)
		assert.Equal(t, *original.Delta, *restored.Delta)
		assert.Nil(t, restored.Value)
	})
}

func TestMetrics_EdgeCases(t *testing.T) {
	t.Run("handles zero values", func(t *testing.T) {
		value := 0.0
		delta := int64(0)
		metric := Metrics{
			ID:    "zero_metric",
			MType: "gauge",
			Value: &value,
			Delta: &delta,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		var restored Metrics
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, 0.0, *restored.Value)
		assert.Equal(t, int64(0), *restored.Delta)
	})

	t.Run("handles negative values", func(t *testing.T) {
		value := -123.45
		delta := int64(-789)
		metric := Metrics{
			ID:    "negative_metric",
			MType: "gauge",
			Value: &value,
			Delta: &delta,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		var restored Metrics
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, -123.45, *restored.Value)
		assert.Equal(t, int64(-789), *restored.Delta)
	})

	t.Run("handles very large values", func(t *testing.T) {
		value := 1.7976931348623157e+308    // близко к максимальному float64
		delta := int64(9223372036854775807) // максимальный int64
		metric := Metrics{
			ID:    "large_metric",
			MType: "gauge",
			Value: &value,
			Delta: &delta,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)

		var restored Metrics
		err = json.Unmarshal(data, &restored)
		require.NoError(t, err)

		assert.Equal(t, value, *restored.Value)
		assert.Equal(t, delta, *restored.Delta)
	})
}
