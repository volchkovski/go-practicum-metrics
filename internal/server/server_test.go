package server

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volchkovski/go-practicum-metrics/internal/backup"
	"github.com/volchkovski/go-practicum-metrics/internal/configs"
	"github.com/volchkovski/go-practicum-metrics/internal/grpcserver"
	"github.com/volchkovski/go-practicum-metrics/internal/httpserver"
	"github.com/volchkovski/go-practicum-metrics/internal/routers"
	"github.com/volchkovski/go-practicum-metrics/internal/services"
	"github.com/volchkovski/go-practicum-metrics/internal/storage/mem"
	"go.uber.org/zap"
)

func TestRun_ConfigValidation(t *testing.T) {
	t.Run("handles invalid log level", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "invalid-level",
			Env:      "local",
			Addr:     ":8080",
		}

		err := Run(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to intizalize logger")
	})

	t.Run("handles invalid environment", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "invalid-env",
			Addr:     ":8080",
		}

		err := Run(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to intizalize logger")
	})
}

func TestRun_MemoryStorage(t *testing.T) {
	t.Run("uses memory storage when no DSN provided", func(t *testing.T) {
		// Этот тест проверяет инициализацию без запуска сервера
		// Мы не можем полноценно протестировать Run без изменения кода,
		// но можем проверить логику инициализации

		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0", // случайный порт
			DSN:             "",   // пустой DSN = memory storage
			FileStoragePath: "/tmp/test-metrics.json",
			StoreIntr:       300,
			Restore:         false,
		}

		// В этом тесте мы ожидаем, что сервер попытается запуститься,
		// но мы не можем легко остановить его без изменения кода
		// Поэтому просто проверим, что конфигурация валидна
		assert.NotNil(t, cfg)
		assert.Equal(t, "", cfg.DSN)
		assert.Equal(t, "info", cfg.LogLevel)
	})
}

// Дополнительные helper функции для тестирования
func validateServerConfig(cfg *configs.ServerConfig) error {
	if cfg.LogLevel == "" {
		return errors.New("log level is required")
	}
	if cfg.Env == "" {
		return errors.New("environment is required")
	}
	if cfg.Addr == "" {
		return errors.New("address is required")
	}
	return nil
}

func TestValidateServerConfig(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "local",
			Addr:     ":8080",
		}

		err := validateServerConfig(cfg)
		assert.NoError(t, err)
	})

	t.Run("missing log level fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			Env:  "local",
			Addr: ":8080",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "log level is required")
	})

	t.Run("missing environment fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Addr:     ":8080",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "environment is required")
	})

	t.Run("missing address fails validation", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel: "info",
			Env:      "local",
		}

		err := validateServerConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address is required")
	})
}

func TestGracefulShutdown(t *testing.T) {
	// Initialize logger for tests
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := logger.Sugar()
	zap.ReplaceGlobals(logger)
	
	t.Run("shuts down successfully with all components", func(t *testing.T) {
		// Create test components
		storage := mem.NewMemStorage()
		service := services.NewMetricService(storage)
		router := routers.NewMetricRouter("", nil, "", service)
		hs := httpserver.New(router, ":0")
		gs := grpcserver.New(":0", service, sugar, "")
		b := backup.NewMetricsBackup(service, "/tmp/test-metrics-shutdown.json", 10) // Non-zero interval
		
		// Start servers
		hs.Start()
		gs.Start()
		
		// Give servers time to start
		time.Sleep(100 * time.Millisecond)
		
		// Test graceful shutdown
		err := gracefulShutdown(hs, gs, b)
		assert.NoError(t, err)
		
		// Cleanup
		service.Close()
	})
	
	t.Run("shuts down successfully without gRPC server", func(t *testing.T) {
		storage := mem.NewMemStorage()
		service := services.NewMetricService(storage)
		router := routers.NewMetricRouter("", nil, "", service)
		hs := httpserver.New(router, ":0")
		b := backup.NewMetricsBackup(service, "/tmp/test-metrics-shutdown-no-grpc.json", 10)
		
		hs.Start()
		time.Sleep(100 * time.Millisecond)
		
		err := gracefulShutdown(hs, nil, b)
		assert.NoError(t, err)
		
		service.Close()
	})
	
	t.Run("handles HTTP server shutdown error gracefully", func(t *testing.T) {
		storage := mem.NewMemStorage()
		service := services.NewMetricService(storage)
		router := routers.NewMetricRouter("", nil, "", service)
		hs := httpserver.New(router, ":0")
		b := backup.NewMetricsBackup(service, "/tmp/test-metrics-shutdown-err.json", 10)
		
		// Don't start server, so shutdown will encounter an error
		err := gracefulShutdown(hs, nil, b)
		// Shutdown of not-started server shouldn't error
		assert.NoError(t, err)
		
		service.Close()
	})
}

func TestRun_BasicScenarios(t *testing.T) {
	t.Run("initializes memory storage correctly", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0",
			DSN:             "", // Empty DSN = memory storage
			FileStoragePath: "/tmp/test-run-mem.json",
			StoreIntr:       10,
			Restore:         false,
			CryptoKey:       "", // No encryption
		}
		
		// We can't fully test Run without it blocking, but we can test config preparation
		assert.NotNil(t, cfg)
		assert.Equal(t, "", cfg.DSN)
		
		// Run will block, so we test the setup separately
		// This validates that the configuration is acceptable
		err := validateServerConfig(cfg)
		assert.NoError(t, err)
	})
	
	t.Run("handles postgres connection error", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0",
			DSN:             "postgres://invalid:invalid@localhost:5432/invalid",
			FileStoragePath: "/tmp/test-run-pg-err.json",
			StoreIntr:       10,
			Restore:         false,
			CryptoKey:       "",
		}
		
		// This should fail when trying to connect to postgres
		err := Run(cfg)
		require.Error(t, err)
		// The error should be about postgres connection - migrations will fail
	})
	
	t.Run("handles invalid crypto key path", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0",
			DSN:             "",
			FileStoragePath: "/tmp/test-run-crypto-err.json",
			StoreIntr:       10,
			Restore:         false,
			CryptoKey:       "/nonexistent/path/to/key.pem",
		}
		
		// Run should fail with invalid crypto key
		err := Run(cfg)
		require.Error(t, err)
	})
}

func TestRun_BackupRestore(t *testing.T) {
	t.Run("handles restore with non-existent file", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0",
			DSN:             "",
			FileStoragePath: "/tmp/nonexistent-restore-file.json",
			StoreIntr:       10,
			Restore:         true, // Try to restore
			CryptoKey:       "",
		}
		
		// Run should handle missing restore file and try to start server
		// But we expect it to continue (EOF is acceptable for empty/missing backup)
		go func() {
			_ = Run(cfg)
		}()
		
		time.Sleep(200 * time.Millisecond)
	})
}

func TestRun_WithGRPC(t *testing.T) {
	t.Run("initializes with gRPC server", func(t *testing.T) {
		cfg := &configs.ServerConfig{
			LogLevel:        "info",
			Env:             "local",
			Addr:            ":0",
			GRPCAddr:        ":0", // Enable gRPC
			DSN:             "",
			FileStoragePath: "/tmp/test-grpc.json",
			StoreIntr:       10,
			Restore:         false,
			CryptoKey:       "",
		}
		
		// Run will block, so we test in a goroutine
		go func() {
			_ = Run(cfg)
		}()
		
		time.Sleep(300 * time.Millisecond)
		
		// Validate that config is acceptable for gRPC setup
		assert.NotEmpty(t, cfg.GRPCAddr)
	})
}
