package configs

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentConfig_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal valid JSON with durations as nanoseconds", func(t *testing.T) {
		jsonData := `{
			"address": "localhost:8080",
			"grpc_address": "localhost:3200",
			"use_grpc": false,
			"report_interval": 10000000000,
			"poll_interval": 2000000000,
			"key": "test-key",
			"rate_limit": 5,
			"crypto_key": "/path/to/key.pem"
		}`

		var cfg AgentConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, "localhost:8080", cfg.ServerAddr)
		assert.Equal(t, "localhost:3200", cfg.GRPCServerAddr)
		assert.Equal(t, false, cfg.UseGRPC)
		assert.Equal(t, 10, cfg.ReportIntr)
		assert.Equal(t, 2, cfg.PollIntr)
		assert.Equal(t, "test-key", cfg.Key)
		assert.Equal(t, 5, cfg.RateLimit)
		assert.Equal(t, "/path/to/key.pem", cfg.CryptoKey)
	})

	t.Run("unmarshal JSON with minute duration", func(t *testing.T) {
		jsonData := `{
			"address": "localhost:8080",
			"report_interval": 60000000000,
			"poll_interval": 30000000000
		}`

		var cfg AgentConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, 60, cfg.ReportIntr)
		assert.Equal(t, 30, cfg.PollIntr)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		jsonData := `{invalid json}`

		var cfg AgentConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		assert.Error(t, err)
	})

	t.Run("handles zero duration", func(t *testing.T) {
		jsonData := `{
			"address": "localhost:8080",
			"report_interval": 0,
			"poll_interval": 0
		}`

		var cfg AgentConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, 0, cfg.ReportIntr)
		assert.Equal(t, 0, cfg.PollIntr)
	})
}

func TestServerConfig_UnmarshalJSON(t *testing.T) {
	t.Run("unmarshal valid server JSON with duration", func(t *testing.T) {
		jsonData := `{
			"address": ":8080",
			"grpc_address": ":3200",
			"store_interval": 300000000000,
			"store_file": "/tmp/metrics.json",
			"restore": true,
			"database_dsn": "postgres://user:pass@localhost/db",
			"crypto_key": "/path/to/private.pem",
			"trusted_subnet": "192.168.1.0/24"
		}`

		var cfg ServerConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, ":8080", cfg.Addr)
		assert.Equal(t, ":3200", cfg.GRPCAddr)
		assert.Equal(t, 300, cfg.StoreIntr)
		assert.Equal(t, "/tmp/metrics.json", cfg.FileStoragePath)
		assert.Equal(t, true, cfg.Restore)
		assert.Equal(t, "postgres://user:pass@localhost/db", cfg.DSN)
		assert.Equal(t, "/path/to/private.pem", cfg.CryptoKey)
		assert.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet)
	})

	t.Run("unmarshal JSON with minute duration", func(t *testing.T) {
		jsonData := `{
			"address": ":8080",
			"store_interval": 300000000000
		}`

		var cfg ServerConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, 300, cfg.StoreIntr)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		jsonData := `{invalid server json}`

		var cfg ServerConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		assert.Error(t, err)
	})

	t.Run("handles zero duration", func(t *testing.T) {
		jsonData := `{
			"address": ":8080",
			"store_interval": 0
		}`

		var cfg ServerConfig
		err := json.Unmarshal([]byte(jsonData), &cfg)
		require.NoError(t, err)

		assert.Equal(t, 0, cfg.StoreIntr)
	})
}

func TestProcessConfigFile_Agent(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("loads config from file with -c flag", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agent_config.json")
		configJSON := `{
			"address": "file.example.com:8080",
			"report_interval": 15000000000,
			"poll_interval": 3000000000
		}`
		err := os.WriteFile(configPath, []byte(configJSON), 0644)
		require.NoError(t, err)

		os.Args = []string{"test", "-c", configPath}

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		require.NoError(t, err)

		assert.Equal(t, "file.example.com:8080", cfg.ServerAddr)
		assert.Equal(t, 15, cfg.ReportIntr)
		assert.Equal(t, 3, cfg.PollIntr)
		assert.Empty(t, args) // Should return empty args after config file
	})

	t.Run("loads config from file with -config flag", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agent_config2.json")
		configJSON := `{
			"address": "config.example.com:9090",
			"report_interval": 20000000000,
			"poll_interval": 5000000000
		}`
		err := os.WriteFile(configPath, []byte(configJSON), 0644)
		require.NoError(t, err)

		os.Args = []string{"test", "-config", configPath}

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		require.NoError(t, err)

		assert.Equal(t, "config.example.com:9090", cfg.ServerAddr)
		assert.Equal(t, 20, cfg.ReportIntr)
		assert.Equal(t, 5, cfg.PollIntr)
		assert.Empty(t, args)
	})

	t.Run("loads config from CONFIG env var", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "agent_env_config.json")
		configJSON := `{
			"address": "env.example.com:7070",
			"report_interval": 25000000000,
			"poll_interval": 7000000000
		}`
		err := os.WriteFile(configPath, []byte(configJSON), 0644)
		require.NoError(t, err)

		os.Args = []string{"test"}
		err = os.Setenv("CONFIG", configPath)
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("CONFIG") }()

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		require.NoError(t, err)

		assert.Equal(t, "env.example.com:7070", cfg.ServerAddr)
		assert.Equal(t, 25, cfg.ReportIntr)
		assert.Equal(t, 7, cfg.PollIntr)
		assert.Empty(t, args)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-c", "/non/existent/path/config.json"}

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		assert.Error(t, err)
		assert.Nil(t, args)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("returns error for invalid JSON in file", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(configPath, []byte("{invalid json}"), 0644)
		require.NoError(t, err)

		os.Args = []string{"test", "-c", configPath}

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		assert.Error(t, err)
		assert.Nil(t, args)
	})

	t.Run("handles no config file specified", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-a", "localhost:8080"}

		cfg := &AgentConfig{}
		args, err := processConfigFile(cfg)
		require.NoError(t, err)

		// Should return remaining args
		assert.Equal(t, []string{"-a", "localhost:8080"}, args)
	})
}

func TestProcessConfigFile_Server(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("loads server config from file", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "server_config.json")
		configJSON := `{
			"address": ":9090",
			"grpc_address": ":3300",
			"store_interval": 600000000000,
			"store_file": "/tmp/server_metrics.json",
			"restore": true,
			"database_dsn": "postgres://test:test@localhost/testdb",
			"trusted_subnet": "10.0.0.0/8"
		}`
		err := os.WriteFile(configPath, []byte(configJSON), 0644)
		require.NoError(t, err)

		os.Args = []string{"test", "-c", configPath}

		cfg := &ServerConfig{}
		args, err := processConfigFile(cfg)
		require.NoError(t, err)

		assert.Equal(t, ":9090", cfg.Addr)
		assert.Equal(t, ":3300", cfg.GRPCAddr)
		assert.Equal(t, 600, cfg.StoreIntr)
		assert.Equal(t, "/tmp/server_metrics.json", cfg.FileStoragePath)
		assert.Equal(t, true, cfg.Restore)
		assert.Equal(t, "postgres://test:test@localhost/testdb", cfg.DSN)
		assert.Equal(t, "10.0.0.0/8", cfg.TrustedSubnet)
		assert.Empty(t, args)
	})
}

func TestParseConfigFile(t *testing.T) {
	t.Run("parses valid config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test_config.json")
		configJSON := `{
			"address": "parse.example.com:8080",
			"report_interval": 10000000000,
			"poll_interval": 2000000000
		}`
		err := os.WriteFile(configPath, []byte(configJSON), 0644)
		require.NoError(t, err)

		cfg := &AgentConfig{}
		err = parseConfigFile(cfg, configPath)
		require.NoError(t, err)

		assert.Equal(t, "parse.example.com:8080", cfg.ServerAddr)
		assert.Equal(t, 10, cfg.ReportIntr)
		assert.Equal(t, 2, cfg.PollIntr)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		cfg := &AgentConfig{}
		err := parseConfigFile(cfg, "/non/existent/file.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open config file")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(configPath, []byte("not valid json"), 0644)
		require.NoError(t, err)

		cfg := &AgentConfig{}
		err = parseConfigFile(cfg, configPath)
		assert.Error(t, err)
	})
}
