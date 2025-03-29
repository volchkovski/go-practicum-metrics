package agent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentNew(t *testing.T) {
	a := New("localhost:8080", 10, 2)
	require.NotNil(t, a)
	assert.NotNil(t, a.memStats)
	assert.Equal(t, "localhost:8080", a.serverAddr)
	assert.Equal(t, 10*time.Second, a.repIntr)
	assert.Equal(t, 2*time.Second, a.pollIntr)
	assert.Equal(t, a.pollCount, 0)
}

func TestGetRandomInt(t *testing.T) {
	x := getRandomInt()
	time.Sleep(1 * time.Second)
	y := getRandomInt()
	assert.NotEqual(t, x, y)
}
