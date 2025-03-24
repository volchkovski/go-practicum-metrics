package storage

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	s := NewMemStorage()
	require.NotNil(t, s)
	require.NotNil(t, s.GS)
	require.NotNil(t, s.CS)
	assert.NotNil(t, s.GS.Gauges)
	assert.NotNil(t, s.CS.Counters)
}

func TestStorage(t *testing.T) {
	tests := []struct {
		name    string
		gauge   float64
		counter int64
		s       Storage
	}{
		{
			name:    "MemStorage",
			gauge:   10,
			counter: 1,
			s:       NewMemStorage(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.s.WriteGauge("gaugeName", tt.gauge)
			assert.Nil(t, err)
			err = tt.s.WriteCounter("counterName", tt.counter)
			assert.Nil(t, err)
		})
	}
}
