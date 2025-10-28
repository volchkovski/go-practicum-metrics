package configs

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerConfig(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("default server config", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test"}

		config, err := NewServerConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, ":8080", config.Addr)
		assert.Equal(t, 300, config.StoreIntr)
		assert.Equal(t, "./metrics.json", config.FileStoragePath)
		assert.Equal(t, false, config.Restore)
		assert.Equal(t, "info", config.LogLevel)
		assert.Equal(t, "local", config.Env)
		assert.Equal(t, "", config.DSN)
		assert.Equal(t, "", config.Key)
	})

	t.Run("server config with flags", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{
			"test",
			"-a", ":9090",
			"-i", "600",
			"-f", "/tmp/test-metrics.json",
			"-r", "true",
		}

		config, err := NewServerConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, ":9090", config.Addr)
		assert.Equal(t, 600, config.StoreIntr)
		assert.Equal(t, "/tmp/test-metrics.json", config.FileStoragePath)
		assert.Equal(t, true, config.Restore)
		// Other fields will be defaults
	})

	t.Run("server config with env vars", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test"}

		// Set environment variables
		var err error
		err = os.Setenv("ADDRESS", ":7070")
		require.NoError(t, err)
		err = os.Setenv("STORE_INTERVAL", "120")
		require.NoError(t, err)
		err = os.Setenv("FILE_STORAGE_PATH", "/var/metrics.json")
		require.NoError(t, err)
		err = os.Setenv("RESTORE", "true")
		require.NoError(t, err)
		err = os.Setenv("LOG_LEVEL", "error")
		require.NoError(t, err)
		err = os.Setenv("ENVIRONMENT", "prod")
		require.NoError(t, err)
		err = os.Setenv("DATABASE_DSN", "postgres://env:pass@localhost/envdb")
		require.NoError(t, err)
		err = os.Setenv("KEY", "env-server-key")
		require.NoError(t, err)

		defer func() {
			err := os.Unsetenv("ADDRESS")
			require.NoError(t, err)
			err = os.Unsetenv("STORE_INTERVAL")
			require.NoError(t, err)
			err = os.Unsetenv("FILE_STORAGE_PATH")
			require.NoError(t, err)
			err = os.Unsetenv("RESTORE")
			require.NoError(t, err)
			err = os.Unsetenv("LOG_LEVEL")
			require.NoError(t, err)
			err = os.Unsetenv("ENVIRONMENT")
			require.NoError(t, err)
			err = os.Unsetenv("DATABASE_DSN")
			require.NoError(t, err)
			err = os.Unsetenv("KEY")
			require.NoError(t, err)
		}()

		config, err := NewServerConfig()
		require.NoError(t, err)

		assert.NotNil(t, config)
		assert.Equal(t, ":7070", config.Addr)
		assert.Equal(t, 120, config.StoreIntr)
		assert.Equal(t, "/var/metrics.json", config.FileStoragePath)
		assert.Equal(t, true, config.Restore)
		assert.Equal(t, "error", config.LogLevel)
		assert.Equal(t, "prod", config.Env)
		assert.Equal(t, "postgres://env:pass@localhost/envdb", config.DSN)
		assert.Equal(t, "env-server-key", config.Key)
	})
}

func TestParseServerFlags(t *testing.T) {
	// Save and restore original command line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("parse server flags", func(t *testing.T) {
		// Reset flags for clean test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		args := []string{
			"-a", ":3000",
			"-i", "60",
			"-f", "/tmp/server-metrics.json",
			"-r", "true",
		}

		config := &ServerConfig{}

		assert.NotPanics(t, func() {
			err := parseServerConfigFields(config, args)
			assert.NoError(t, err)
		})

		assert.Equal(t, ":3000", config.Addr)
		assert.Equal(t, 60, config.StoreIntr)
		assert.Equal(t, "/tmp/server-metrics.json", config.FileStoragePath)
		assert.Equal(t, true, config.Restore)
		// Other fields will be defaults
	})
}
