package pg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadQueries(t *testing.T) {
	t.Run("loads queries successfully", func(t *testing.T) {
		err := loadQueries()
		require.NoError(t, err)

		// Verify that queries were loaded
		assert.NotEmpty(t, q.InsertGauge)
		assert.NotEmpty(t, q.InsertCounter)
		assert.NotEmpty(t, q.SelectGaugeValue)
		assert.NotEmpty(t, q.SelectCounterValue)
		assert.NotEmpty(t, q.SelectGauges)
		assert.NotEmpty(t, q.SelectCounters)
	})

	t.Run("subsequent calls return no error", func(t *testing.T) {
		// sync.Once ensures this doesn't reload
		err := loadQueries()
		assert.NoError(t, err)
	})
}

func TestLoadQuery(t *testing.T) {
	t.Run("loads existing query", func(t *testing.T) {
		query, err := loadQuery("insert_gauge")
		require.NoError(t, err)
		assert.NotEmpty(t, query)
		assert.Contains(t, query, "INSERT") // Assuming it's an INSERT query
	})

	t.Run("returns error for non-existent query", func(t *testing.T) {
		_, err := loadQuery("nonexistent_query")
		assert.Error(t, err)
	})
}

