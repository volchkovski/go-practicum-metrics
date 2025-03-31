package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	require.NotNil(t, s)
	assert.NotNil(t, s.gauges)
	assert.NotNil(t, s.counters)
}

func TestMemStorage(t *testing.T) {
	s := NewMemStorage()

	t.Run("test gauges", func(t *testing.T) {
		s.WriteGauge("testg", float64(123))
		require.Contains(t, s.gauges, "testg")
		require.Equal(t, float64(123), s.gauges["testg"])
		s.WriteGauge("testg", float64(321))
		gg, err := s.ReadGauge("testg")
		require.Nil(t, err)
		require.Equal(t, float64(321), gg)
		assert.Equal(t, map[string]float64{"testg": 321}, s.ReadAllGauges())
	})

	t.Run("test counters", func(t *testing.T) {
		s.WriteCounter("testc", int64(1))
		require.Contains(t, s.counters, "testc")
		require.Equal(t, s.counters["testc"], int64(1))
		s.WriteCounter("testc", int64(2))
		cn, err := s.ReadCounter("testc")
		require.Nil(t, err)
		require.Equal(t, int64(3), cn)
		assert.Equal(t, map[string]int64{"testc": 3}, s.ReadAllCounters())
	})
}
