package configs

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConfig(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("default config", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test"}

		config, err := NewAgentConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, "localhost:8080", config.ServerAddr)
		assert.Equal(t, 10, config.ReportIntr)
		assert.Equal(t, 2, config.PollIntr)
		assert.Equal(t, "", config.Key)
		assert.Equal(t, 1, config.RateLimit)
	})

	t.Run("config with flags", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{
			"test",
			"-a", "test.example.com:9090",
			"-r", "5",
			"-p", "1",
			"-k", "test-key",
			"-l", "3",
		}

		config, err := NewAgentConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, "test.example.com:9090", config.ServerAddr)
		assert.Equal(t, 5, config.ReportIntr)
		assert.Equal(t, 1, config.PollIntr)
		assert.Equal(t, "test-key", config.Key)
		assert.Equal(t, 3, config.RateLimit)
	})

	t.Run("config with env vars", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test"}

		// Set environment variables
		var err error
		err = os.Setenv("ADDRESS", "env.example.com:8080")
		require.NoError(t, err)
		err = os.Setenv("REPORT_INTERVAL", "15")
		require.NoError(t, err)
		err = os.Setenv("POLL_INTERVAL", "3")
		require.NoError(t, err)
		err = os.Setenv("KEY", "env-key")
		require.NoError(t, err)
		err = os.Setenv("RATE_LIMIT", "5")
		require.NoError(t, err)

		defer func() {
			err := os.Unsetenv("ADDRESS")
			require.NoError(t, err)
			err = os.Unsetenv("REPORT_INTERVAL")
			require.NoError(t, err)
			err = os.Unsetenv("POLL_INTERVAL")
			require.NoError(t, err)
			err = os.Unsetenv("KEY")
			require.NoError(t, err)
			err = os.Unsetenv("RATE_LIMIT")
			require.NoError(t, err)
		}()

		config, err := NewAgentConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, "env.example.com:8080", config.ServerAddr)
		assert.Equal(t, 15, config.ReportIntr)
		assert.Equal(t, 3, config.PollIntr)
		assert.Equal(t, "env-key", config.Key)
		assert.Equal(t, 5, config.RateLimit)
	})
}

func TestParseAgentFlags(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("parse flags", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		args := []string{
			"-a", "flagtest.com:8080",
			"-r", "20",
			"-p", "5",
			"-k", "flag-key",
			"-l", "10",
		}

		config := &AgentConfig{}

		assert.NotPanics(t, func() {
			parseAgentFlags(config, args)
		})

		assert.Equal(t, "flagtest.com:8080", config.ServerAddr)
		assert.Equal(t, 20, config.ReportIntr)
		assert.Equal(t, 5, config.PollIntr)
		assert.Equal(t, "flag-key", config.Key)
		assert.Equal(t, 10, config.RateLimit)
	})
}
