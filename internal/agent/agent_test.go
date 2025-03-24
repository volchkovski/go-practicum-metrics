package agent

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAgentNew(t *testing.T) {
	a := New()
	require.NotNil(t, a)
	assert.NotNil(t, a.memStats)
	assert.Equal(t, a.pollCount, 0)
}

func TestGetRandomInt(t *testing.T) {
	x := getRandomInt()
	time.Sleep(1 * time.Second)
	y := getRandomInt()
	assert.NotEqual(t, x, y)
}
